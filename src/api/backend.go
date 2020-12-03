package api

import (
	"github.com/apvodney/JumpDir/api/fatalError"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx"

	"bufio"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var GenericSQLError error = errors.New("Error running SQL request")

type randBuf struct {
	idc chan []byte
	err error
	ctx context.Context
}

var rb *randBuf

func init() {
	rb = new(randBuf)
	rb.idc = make(chan []byte)
	var cancel context.CancelFunc
	rb.ctx, cancel = context.WithCancel(context.Background())
	randomWorker := func(rb *randBuf) {
		b := bufio.NewReader(rand.Reader)
		for {
			id := make([]byte, 16)
			_, err := b.Read(id)
			if err != nil {
				rb.err = errors.New("Failed to generate unique ID")
				cancel()
				return
			}
			select {
			case <-rb.ctx.Done():
				return
			case rb.idc <- id:
			}
		}
	}
	go randomWorker(rb)
	go randomWorker(rb)
}

func (rb *randBuf) ID() []byte {
	select {
	case <-rb.ctx.Done():
		panic(rb.err)
	case id := <-rb.idc:
		return id
	}
}

// Takes f, a sql function whose last argument is a bytea value that is meant to take a random unique value.
// The function must return a boolean value indicating the success of the unique insert. The args passed must
// complete all but the last of the arguments required by f. p takes an error type and returns a boolean
// indicating whether the error was handled, if the boolean returned is true, attemptIdInsert returns nil.
func (a *Api) attemptIdInsert(ctx context.Context, p func(error) bool, f string, args ...interface{}) []byte {
	const sel string = "SELECT "
	var (
		qbuf    strings.Builder
		id      []byte
		err     error
		success bool = false

		args2 []interface{} = append(args, id)
	)

	// This works because each argument in the query is `$#, ` except for the last,
	// which is just `$#` but we also need to account for the two parentheses, so it works out.
	qbuf.Grow(len(sel) + len(f) + len(args)*4)
	fmt.Fprintf(&qbuf, "%s%s(", sel, f)
	for i := range args2 {
		fmt.Fprintf(&qbuf, "$%d", i+1)
		if i < len(args2)-1 {
			qbuf.WriteString(", ")
		}
	}
	qbuf.WriteRune(')')
	query := qbuf.String()
	println(query)
	for !success {
		id = rb.ID()
		args2[len(args2)-1] = id
		err = a.sql.QueryRow(ctx, query, args2...).Scan(&success)
		if p != nil && p(err) {
			return nil
		} else if err != nil {
			log.Printf("%#v", err)
			fatalError.Panic(err)
		}
		if ctx.Err() != nil {
			fatalError.Panic(ctx.Err())
		}
	}
	return id
}

// Speculatively inserts user into database, and returns a secret to be delivered to the user
// by email to verify their identity.
func (a *Api) startReg(ctx context.Context, username, email, password string) (secret []byte, err error) {
	if username == "" || email == "" || password == "" {
		return nil, fmt.Errorf("Username (%s), email (%s) and password (%s) must all not be empty",
			username, email, strings.Repeat("*", len(password)))
	}
	passHash, hashAlgo := a.passHash(password)

	secret = a.attemptIdInsert(ctx, nil, "attempt_reg_pending_insert", time.Now().Add(3*24*time.Hour),
		username, email, hashAlgo, passHash)

	return secret, nil
}

func (a *Api) authenticate(ctx context.Context, username, password string) ([]byte, error) {
	var (
		authFail     error = errors.New("Incorrect username or password")
		userID       []byte
		hashAlgo     int32
		passwordHash string
	)
	err := a.sql.QueryRow(ctx, "SELECT hashAlgo, passHash FROM users WHERE username = $1", username).
		Scan(&userID, &hashAlgo, &passwordHash)
	if err == pgx.ErrNoRows {
		return nil, authFail
	} else if err != nil {
		fatalError.Panic(err)
	}
	same := a.passCompare(password, passwordHash, hashAlgo)
	if !same {
		return nil, authFail
	}
	return userID, nil
}

func (a *Api) newToken(ctx context.Context, userID []byte) []byte {
	return a.attemptIdInsert(ctx, nil, "attempt_token_insert", userID, time.Now().Add(31*24*time.Hour))
}

func (a *Api) finishReg(ctx context.Context, secret []byte, email string) (sessionToken []byte, err error) {
	errp := new(error)
	p := func(err error) bool {
		pge, ok := err.(*pgconn.PgError)
		if !ok {
			return false
		}
		switch pge.Code {
		case "P0002":
			*errp = errors.New("Invalid email or secret.")
		case "23505":
			*errp = errors.New("Username already registered.")
		case "ZD000":
			*errp = errors.New("Verification token expired.")
		default:
			return false
		}
		return true
	}

	userID := a.attemptIdInsert(ctx, p, "attempt_finish_reg", secret, email, time.Now())
	if *errp != nil {
		return nil, *errp
	}
	token := a.newToken(ctx, userID)
	return token, nil
}

func (a *Api) getToken(ctx context.Context, username, password string) (token []byte, err error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("Username (%s) and password (%s) must not be empty",
			username, strings.Repeat("*", len(password)))
	}
	userID, err := a.authenticate(ctx, username, password)
	if err != nil {
		return nil, err
	}
	token = a.newToken(ctx, userID)
	return token, nil
}
