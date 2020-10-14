package api

import (
	"github.com/apvodney/JumpDir/debug"

	"database/sql"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rubenv/sql-migrate"

	"context"
	"crypto/rand"
	"fmt"
	"time"
)

func initsql() (error, *pgxpool.Pool) {
	if debug.True {
		dbpool, err := pgxpool.Connect(context.Background(), "host=debug_db user=postgres sslmode=disable pool_max_conns=5")
		defer dbpool.Close()
		if err != nil {
			return fmt.Errorf("Can't connect or invalid config: %w", err), nil
		}
		dbpool.Exec(context.Background(), "DROP DATABASE directory")
		_, err = dbpool.Exec(context.Background(), "CREATE DATABASE directory")
		if err != nil {
			return fmt.Errorf("failed to create DB: %w", err), nil
		}
	}

	db, err := sql.Open("pgx", "host=db user=postgres dbname=directory sslmode=disable")
	defer db.Close()
	if err != nil {
		return fmt.Errorf("Can't connect or invalid config: %w", err), nil
	}

	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}
	_, err = migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("Failure migrating: %w", err), nil
	}

	dbpool, err := pgxpool.Connect(context.Background(), "host=db user=postgres dbname=directory sslmode=disable pool_max_conns=5")
	if err != nil {
		return fmt.Errorf("Can't connect or invalid config: %w", err), nil
	}
	return nil, dbpool
}

type Api struct {
	sql         *pgxpool.Pool
	hashLimiter chan struct{}
}

func Initialize() (error, *Api) {
	a := new(Api)
	err, sql := initsql()
	if err != nil {
		return err, nil
	}
	a.sql = sql
	a.hashLimiter = make(chan struct{}, 32)
	for i := 0; i < cap(a.hashLimiter); i++ {
		a.hashLimiter <- struct{}{}
	}
	return nil, a
}

func uniqueID() (error, []byte) {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		return fmt.Errorf("Error getting unique random ID: %w", err), nil
	}
	return err, id
}

// Speculatively inserts user into database, and returns a secret to be delivered to the user
// by email to verify their identity.
func (a *Api) StartReg(ctx context.Context, email, password string) (error, []byte) {
	var deadline = time.Now().Add(3 * 24 * time.Hour)
	println(deadline.Format(time.UnixDate))

	if email == "" || password == "" {
		return fmt.Errorf("Both username (%s) and password (%s) must not be empty", email, password), nil
	}
	err, passHash, hashAlgo := a.passHash(password)
	if err != nil {
		return err, nil
	}

	// Try repeatedly to insert a new user until we find a unique userid
	var id []byte
	for {
		var err error
		err, id = uniqueID()
		if err != nil {
			return err, nil
		}
		var success bool
		err = a.sql.QueryRow(ctx, "SELECT attempt_start_reg($1, $2, $3, $4)",
			id, email, hashAlgo, passHash).Scan(&success)
		if err != nil {
			return err, nil
		}
		if success {
			break
		}
		if ctx.Err() != nil {
			return ctx.Err(), nil
		}
	}

	// Try repeatedly to insert an authentication token.
	var secret []byte
	for {
		var err error
		err, secret = uniqueID()
		if err != nil {
			return err, nil
		}
		var success bool
		err = a.sql.QueryRow(ctx, "SELECT attempt_reg_secret_insert($1, $2, $3)",
			secret, id, deadline).Scan(&success)
		if err != nil {
			return err, nil
		}
		if success {
			break
		}
		if ctx.Err() != nil {
			return ctx.Err(), nil
		}
	}
	return nil, secret
}

// func FinishReg(ctx context.Context, email string, secret []byte) bool {
//
// }
