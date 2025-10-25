package store

import (
	"context"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
)

type IPrincipalStore interface {
	FindUserByID(ctx context.Context, id int64) (*model.User, error)
}
