package fit

import (
	"context"

	"github.com/easysoft/gitfox-community/app/service/utils"
)

// Interface fit 服务接口
type Interface interface {
	// GetBranchPermission 通过外部 DevOps 系统获取分支权限
	GetBranchPermission(ctx context.Context, repoID uint, branchName string) (*utils.BranchPermissionResult, error)
}
