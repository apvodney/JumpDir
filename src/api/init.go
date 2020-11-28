package api

import (
	"github.com/apvodney/JumpDir/debug"

	"database/sql"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rubenv/sql-migrate"

	"context"
	"fmt"
	"time"
)

type Api struct {
	sql interface {
		Query(context.Context, string, ...interface{}) (pgx.Rows, error)
		QueryRow(context.Context, string, ...interface{}) pgx.Row
		Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	}
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

func initsql() (error, *pgxpool.Pool) {
	if debug.True {
		var dbpool *pgxpool.Pool
		for {
			var err error
			dbpool, err = pgxpool.Connect(context.Background(), "host=debug_db user=postgres sslmode=disable pool_max_conns=5")
			if err != nil {
				fmt.Printf("Can't connect or invalid config: %s\n", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			break
		}
		_, err := dbpool.Exec(context.Background(), "CREATE DATABASE directory")
		if err != nil {
			return fmt.Errorf("failed to create DB: %w", err), nil
		}
		dbpool.Close()
	}
	var db *sql.DB
	for {
		var err error
		db, err = sql.Open("pgx", "host=db user=postgres dbname=directory sslmode=disable")
		if err != nil {
			fmt.Printf("Can't connect or invalid config: %s\n", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}
	defer db.Close()

	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}
	_, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("Failure migrating: %w", err), nil
	}

	for {
		dbpool, err := pgxpool.Connect(context.Background(), "host=db user=postgres dbname=directory sslmode=disable pool_max_conns=5")
		if err != nil {
			fmt.Printf("Can't connect or invalid config: %s\n", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		return nil, dbpool
	}
}

func (a *Api) Copy() *Api {
	napi := new(Api)
	*napi = *a
	return napi
}
