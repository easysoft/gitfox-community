// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package branch

import (
	"context"

	"github.com/easysoft/gitfox-community/app/model/auth"
	"github.com/easysoft/gitfox-community/app/model/branch"
	"github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/gin-gonic/gin"
)

type Interface interface {
	Create(ctx context.Context, req branch.ReqForCreate, session *auth.Session, repo *repo.Repo) (*branch.Branch, error)
	List(ctx context.Context, req branch.ReqFilter, repo *repo.Repo) ([]branch.Branch, bool, error)
	ListPrefixes(ctx context.Context, repo repo.Repo) ([]branch.Prefix, error)
	Delete(ctx context.Context, name string, session *auth.Session, repo *repo.Repo) error

	// BranchType 相关接口
	BatchCreateBranchType(ctx context.Context, reqs []branch.ReqForCreateBranchType, session *auth.Session, repoID uint) ([]branch.BranchType, error)
	UpdateBranchType(ctx context.Context, id uint, req branch.ReqForUpdateBranchType, session *auth.Session, repoID uint) (*branch.BranchType, error)
	GetBranchType(ctx context.Context, id uint, repoID uint) (*branch.BranchType, error)
	DeleteBranchType(ctx context.Context, id uint, repoID uint) error
	ListBranchTypes(ctx context.Context, filter branch.ReqFilterBranchType, repoID uint) (*database.PageResult[branch.BranchType], error)
}

// IArchiveService 分支归档/激活服务接口。
//
// 分支归档是付费功能。社区版绑定 NoopArchiveService，付费版绑定其自身实现。
type IArchiveService interface {
	BatchArchive(c *gin.Context, ctx context.Context, req branch.ReqForArchive, session *auth.Session, repo *repo.Repo) []branch.ArchiveResult
	BatchActivate(c *gin.Context, ctx context.Context, req branch.ReqForArchive, session *auth.Session, repo *repo.Repo) []branch.ActivateResult
}
