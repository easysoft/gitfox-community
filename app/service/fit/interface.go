// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

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
