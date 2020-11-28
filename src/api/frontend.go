package api

import (
	"context"
)

func (a *Api) serializableRetry(func() error) {
}

func (a *Api) StartReg(ctx context.Context, username, email, password string) (secret []byte, err error) {
	return a.startReg(ctx, username, email, password)
}

func (a *Api) FinishReg(ctx context.Context, secret []byte, email string) (sessionToken []byte, err error) {
	return a.finishReg(ctx, secret, email)
}

func (a *Api) GetToken(ctx context.Context, username, password string) (token []byte, err error) {
	return a.getToken(ctx, username, password)
}
