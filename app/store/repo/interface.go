package repo

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	FindByRef(ctx context.Context, ref string) (*repo.Repository, error)
}
