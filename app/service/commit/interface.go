package commit

import (
	"context"
	"io"

	commitModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	Create(ctx context.Context, req commitModel.ReqForCreate) (*commitModel.RespForCreate, error)
	List(ctx context.Context, req commitModel.ReqFilter, repo repoModel.Repo, gitref string) (commitModel.RespForList, error)
	Get(ctx context.Context, commitSHA string, repo repoModel.Repo) (*commitModel.Commit, error)
	GetDivergence(ctx context.Context, req commitModel.ReqForDivergence, repo repoModel.Repo) ([]commitModel.CommitDivergence, error)
	GetDiff(ctx context.Context, commitSHA string, repo repoModel.Repo, w io.Writer) error
}
