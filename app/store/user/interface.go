package user

import (
	"context"

	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
)

type Interface interface {
	FindByID(ctx context.Context, id int64) (*principal.User, error)
	FindByAccount(ctx context.Context, account string) (*principal.User, error)
	FindByEmail(ctx context.Context, email string) (*principal.User, error)
}

type CacheInterface interface {
	Find(ctx context.Context, account string) (*principal.User, error)
	FindMany(ctx context.Context, accounts []string) ([]*principal.User, error)
}
