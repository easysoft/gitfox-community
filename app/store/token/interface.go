package token

import (
	"context"
	"time"

	"github.com/easysoft/gitfox-community/app/model/auth"
	principal "github.com/easysoft/gitfox-community/app/model/user"
)

type Interface interface {
	GetByID(ctx context.Context, id int64) (*Token, error)
	GetByIdentifier(ctx context.Context, account string, identifier string) (*Token, error)
	GetDevopsToken(ctx context.Context) (*Entry, error)
	SetDevopsToken(ctx context.Context) error
	UpdateToken(ctx context.Context, id int64, token string) error

	CreateUserPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error)
	CreateAppPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error)
}
