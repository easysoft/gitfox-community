package principal

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
)

type IPrincipalStore interface {
	FindUserByID(ctx context.Context, id int64) (*principal.User, error)
	FindUserByUID(ctx context.Context, uid string) (*principal.User, error)
}
