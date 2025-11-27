package repo

import (
	"context"
	"io"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	Create(ctx context.Context, session *auth.Session, req *repoModel.CreateRepoReq) (*repoModel.Repo, error)
	GetByID(ctx context.Context, session *auth.Session, id uint) (*repoModel.Repo, error)
	GetRepoByIdentifier(ctx context.Context, spaceId uint, identifier string, allowedStates []constant.RepoState) (*repoModel.Repo, error)
	GetRepoCheckAccess(ctx context.Context, session *auth.Session, spaceId uint, identifier string, allowedPermissions []constant.Permission) (*repoModel.Repo, error)
	Update(ctx context.Context, req *repoModel.ReqForUpdate, repo repoModel.Repo, session *auth.Session) (*repoModel.Repo, error)
	GetDiff(ctx context.Context, req repoModel.ReqForDiff, repo repoModel.Repo, w io.Writer) error
	Paths(ctx context.Context, req repoModel.PathsInput, repo repoModel.Repo) (*repoModel.PathsOutput, error)
	PathsDetails(ctx context.Context, req repoModel.PathsDetailsInput, repo repoModel.Repo) (*repoModel.PathsDetailsOutput, error)
	Delete(ctx context.Context, session *auth.Session, id uint) error
}
