package branch

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"github.com/gin-gonic/gin"
)

type Interface interface {
	Create(ctx context.Context, req branch.ReqForCreate, session *auth.Session, repo *repo.Repo) (*branch.Branch, error)
	List(ctx context.Context, req branch.ReqFilter, repo *repo.Repo) ([]branch.Branch, bool, error)
	ListPrefixes(ctx context.Context, repo repo.Repo) ([]branch.Prefix, error)
	Delete(ctx context.Context, name string, session *auth.Session, repo *repo.Repo) error
	BatchArchive(c *gin.Context, ctx context.Context, req branch.ReqForArchive, session *auth.Session, repo *repo.Repo) []branch.ArchiveResult
	BatchActivate(c *gin.Context, ctx context.Context, req branch.ReqForArchive, session *auth.Session, repo *repo.Repo) []branch.ActivateResult

	// BranchType 相关接口
	BatchCreateBranchType(ctx context.Context, reqs []branch.ReqForCreateBranchType, session *auth.Session, repoID uint) ([]branch.BranchType, error)
	UpdateBranchType(ctx context.Context, id uint, req branch.ReqForUpdateBranchType, session *auth.Session, repoID uint) (*branch.BranchType, error)
	GetBranchType(ctx context.Context, id uint, repoID uint) (*branch.BranchType, error)
	DeleteBranchType(ctx context.Context, id uint, repoID uint) error
	ListBranchTypes(ctx context.Context, filter branch.ReqFilterBranchType, repoID uint) (*database.PageResult[branch.BranchType], error)
}
