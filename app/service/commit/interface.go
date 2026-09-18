package commit

import (
	"context"
	"io"

	"github.com/easysoft/gitfox-community/app/model/auth"
	"github.com/easysoft/gitfox-community/app/model/commit"
	"github.com/easysoft/gitfox-community/app/model/repo"
)

type Interface interface {
	Create(ctx context.Context, req commit.ReqForCreate, session *auth.Session, repo *repo.Repo) (*commit.RespForCreate, error)
	List(ctx context.Context, req commit.ReqFilter, repo repo.Repo, gitref string) (commit.RespForList, error)
	Get(ctx context.Context, commitSHA string, repo repo.Repo) (*commit.Commit, error)
	GetDivergence(ctx context.Context, req commit.ReqForDivergence, repo repo.Repo) ([]commit.CommitDivergence, error)
	GetDiff(ctx context.Context, commitSHA string, repo repo.Repo, w io.Writer) error
}
