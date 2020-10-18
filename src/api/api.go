package api

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var GenericSQLError error = errors.New("Error running SQL request")

func uniqueID() (error, []byte) {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		log.Println(err.Error())
		return errors.New("Failed to generate user ID"), nil
	}
	return err, id
}

// Takes a query and an incomplete argument list where the last argument is missing, and needs to be a
// random bytea value. The query must return a boolean indicating whether the record was successfully
// inserted (true) or not (false). Returns the byte slice that succeeded.
func (a *Api) attemptIdInsert(ctx context.Context, query string, args ...interface{}) (error, []byte) {
	var (
		id      []byte
		err     error
		success bool = false
	)
	for !success {
		err, id = uniqueID()
		if err != nil {
			return err, nil
		}
		args2 := append(args, id)
		err = a.sql.QueryRow(ctx, query, args2...).Scan(&success)
		if err != nil {
			log.Println(err.Error())
			return GenericSQLError, nil
		}
		if ctx.Err() != nil {
			return ctx.Err(), nil
		}
	}
	return nil, id
}

// Speculatively inserts user into database, and returns a secret to be delivered to the user
// by email to verify their identity.
func (a *Api) StartReg(ctx context.Context, username, email, password string) (error, []byte) {
	if username == "" || email == "" || password == "" {
		return fmt.Errorf("Username (%s), email (%s) and password (%s) must all not be empty",
			username, email, password), nil
	}
	err, passHash, hashAlgo := a.passHash(password)
	if err != nil {
		return err, nil
	}

	err, id := a.attemptIdInsert(ctx, "SELECT attempt_start_reg($1, $2, $3, $4, $5)",
		username, email, hashAlgo, passHash)
	if err != nil {
		return err, nil
	}

	err, secret := a.attemptIdInsert(ctx, "SELECT attempt_reg_secret_insert($1, $2, $3)",
		id, time.Now().Add(3*24*time.Hour))
	if err != nil {
		return err, nil
	}
	return nil, secret
}

func (a *Api) authenticate(ctx context.Context, username, password string) (error, []byte) {
	var (
		userID       []byte
		hashAlgo     int32
		passwordHash string
	)
	err := a.sql.QueryRow(ctx, "SELECT hashAlgo, passHash FROM users WHERE username = $1", username).
		Scan(&userID, &hashAlgo, &passwordHash)
	if err != nil {
		log.Println(err.Error())
		return GenericSQLError, nil
	}
	err, same := a.passCompare(password, passwordHash, hashAlgo)
	if err != nil {
		log.Println(err.Error())
		return errors.New("Failure to compare passwords"), nil
	}
	if !same {
		return errors.New("Incorrect username or password"), nil
	}
	return nil, userID
}

func (a *Api) newToken(ctx context.Context, userID []byte) (error, []byte) {
	err, token := a.attemptIdInsert(ctx, "SELECT attempt_token_insert($1, $2, $3)",
		userID, time.Now().Add(31*24*time.Hour))
	if err != nil {
		log.Println(err.Error())
		return GenericSQLError, nil
	}
	return nil, token
}

func (a *Api) FinishReg(ctx context.Context, secret []byte, email string) (err error, sessionToken []byte) {
	var userID []byte
	err = a.sql.QueryRow(ctx, "SELECT finish_reg($1, $2, $3)",
		secret, email, time.Now()).Scan(&userID)
	if err != nil {
		log.Println(err.Error())
		return GenericSQLError, nil
	}
	if userID == nil {
		return errors.New("Invalid email or secret"), nil
	}
	err, token := a.newToken(ctx, userID)
	if err != nil {
		return err, nil
	}
	return nil, token
}

func (a *Api) GetToken(ctx context.Context, username, password string) (error, []byte) {
	err, userID := a.authenticate(ctx, username, password)
	if err != nil {
		return err, nil
	}
	if userID == nil {
		return errors.New("Incorrect username or password."), nil
	}
	err, token := a.newToken(ctx, userID)
	if err != nil {
		log.Println(err.Error())
		return errors.New("Failure acquiring token"), nil
	}
	return nil, token
}
