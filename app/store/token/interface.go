package token

import (
	"context"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
)

type Interface interface {
	GetByID(ctx context.Context, id int64) (*Token, error)
	GetByIdentifier(ctx context.Context, identifier string) (*Token, error)
	UpdateToken(ctx context.Context, id int64, token string) error

	CreateUserPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error)
	CreateAppPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error)
}
