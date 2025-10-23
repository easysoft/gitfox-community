package ssh

import (
	"context"

	"github.com/gliderlabs/ssh"
)

type ctxSessionKey struct{}

func WithContext(ctx context.Context, session ssh.Session) context.Context {
	return context.WithValue(ctx, ctxSessionKey{}, session)
}

func FromContext(ctx context.Context) ssh.Session {
	obj, ok := ctx.Value(ctxSessionKey{}).(ssh.Session)
	if !ok {
		return nil
	}
	return obj
}
