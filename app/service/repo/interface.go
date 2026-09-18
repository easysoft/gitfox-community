package repo

import (
	"context"
	"io"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/app/model/auth"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	svnModel "github.com/easysoft/gitfox-community/app/model/svn"
	gitAPI "github.com/easysoft/gitfox-community/pkg/git/api"
)

type Interface interface {
	Create(ctx context.Context, session *auth.Session, req *repoModel.CreateRepoReq) (*repoModel.Repo, error)
	GetByID(ctx context.Context, session *auth.Session, id uint) (*repoModel.Repo, error)
	GetByIdentifier(ctx context.Context, spaceID uint, identifier string, allowedStates []string) (*repoModel.Repo, error)
	GetRepoCheckAccess(ctx context.Context, session *auth.Session, spaceID uint, identifier string, allowedPermissions []constant.Permission) (*repoModel.Repo, error)
	GetRepoCheckAccessByID(ctx context.Context, session *auth.Session, repoID uint, allowedPermissions []constant.Permission) (*repoModel.Repo, error)
	Update(ctx context.Context, req *repoModel.ReqForUpdate, repo repoModel.Repo, session *auth.Session) (*repoModel.Repo, error)
	GetDiff(ctx context.Context, req repoModel.ReqForDiff, repo repoModel.Repo, w io.Writer) error
	GetBlame(ctx context.Context, req repoModel.BlameFilter, repo repoModel.Repo, filePath string, w io.Writer) error
	Paths(ctx context.Context, req repoModel.PathsInput, repo repoModel.Repo) (*repoModel.PathsOutput, error)
	PathsDetails(ctx context.Context, req repoModel.PathsDetailsInput, repo repoModel.Repo) (*repoModel.PathsDetailsOutput, error)
	Delete(ctx context.Context, session *auth.Session, id uint) error
	GetContent(ctx context.Context, session *auth.Session, repo *repoModel.Repo, gitRef string, repoPath string, includeLatestCommit bool, flattenDirectories bool) (*repoModel.GetContentOutput, error)
	Archive(ctx context.Context, params gitAPI.ArchiveParams, repo repoModel.Repo, w io.Writer) error
	List(ctx context.Context, session *auth.Session, filter repoModel.ReqFilter) ([]*repoModel.Repo, bool, error)
	ExecSvnCommand(ctx context.Context, repo repoModel.Repo, req svnModel.ReqForExec) (output string, contentType string, err error)
	ExportSvn(ctx context.Context, repo repoModel.Repo, revision string, w io.Writer) (bool, error)
}
