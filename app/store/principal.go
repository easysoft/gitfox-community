package store

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model"
)

type IPrincipalStore interface {
	FindUserByID(ctx context.Context, id int64) (*model.User, error)
}
