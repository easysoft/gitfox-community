package tag

import (
	"context"

	"github.com/easysoft/gitfox-community/app/model/auth"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	tagModel "github.com/easysoft/gitfox-community/app/model/tag"
)

type Interface interface {
	Create(ctx context.Context, req tagModel.ReqForCreate, session *auth.Session, repo *repoModel.Repo) (*tagModel.Tag, error)
	List(ctx context.Context, req tagModel.ReqFilter, repo repoModel.Repo) ([]tagModel.Tag, bool, error)
	Delete(ctx context.Context, name string, session *auth.Session, repo *repoModel.Repo) error
}
