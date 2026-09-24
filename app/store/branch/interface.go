// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package branch

import (
	"context"

	branchModel "github.com/easysoft/gitfox-community/app/model/branch"
	"github.com/easysoft/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

type Interface interface {
	// BranchType 相关接口
	BatchCreateBranchType(ctx context.Context, branchTypes []*BranchType) ([]branchModel.BranchType, error)
	CheckBranchTypePrefix(ctx context.Context, tx *gorm.DB, repo uint, branchTypeID uint, prefixs []string) (bool, error)
	UpdateBranchType(ctx context.Context, branchType *BranchType, id uint) (*branchModel.BranchType, error)
	GetBranchTypeByID(ctx context.Context, id uint) (*branchModel.BranchType, error)
	GetBranchTypesByIDs(ctx context.Context, repoID uint, ids []uint) ([]branchModel.BranchType, error)
	DeleteBranchType(ctx context.Context, id uint) error
	ListBranchTypes(ctx context.Context, repoID uint, filter branchModel.ReqFilterBranchType) (*database.PageResult[branchModel.BranchType], error)
}
