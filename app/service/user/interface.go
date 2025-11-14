package user

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
)

type Interface interface {
	FindUserByUID(ctx context.Context, uid string) (*principal.User, error)
	FindUserByEmail(ctx context.Context, email string) (*principal.User, error)
}
