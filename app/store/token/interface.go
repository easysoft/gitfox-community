package token

import (
	"context"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
)

type Interface interface {
	GetByID(ctx context.Context, id int64) (*Token, error)

	CreateUserSession(ctx context.Context, user *principal.User, tokenIdentifier string) (*auth.Token, error)
	CreateUserPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error)
}
