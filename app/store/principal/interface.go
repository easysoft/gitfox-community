package principal

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
)

type Interface interface {
	FindUserByID(ctx context.Context, id int64) (*principal.User, error)
	FindUserByUID(ctx context.Context, uid string) (*principal.User, error)
	FindUserByEmail(ctx context.Context, email string) (*principal.User, error)
}

type CacheInterface interface {
	Find(ctx context.Context, id int64) (*principal.PrincipalInfo, error)
	FindMany(ctx context.Context, ids []int64) ([]*principal.PrincipalInfo, error)
}
