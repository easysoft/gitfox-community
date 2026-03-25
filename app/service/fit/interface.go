package fit

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
)

// Interface fit 服务接口
type Interface interface {
	// GetBranchPermission 通过外部 DevOps 系统获取分支权限
	GetBranchPermission(ctx context.Context, repoID uint, branchName string) (*utils.BranchPermissionResult, error)
	// GetRepoByName 根据空间 ID 和仓库名获取仓库信息（带权限校验）
	GetRepoByName(ctx context.Context, session *auth.Session, spaceID uint, repoName string) (*repoModel.Repo, error)
}
