// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package branch

import (
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// ***** BranchType(zt_ops_branch_type) 表在devops 库中 *****
// ***** BranchType(zt_ops_branch_type) 表 Repo 字段 存储的是 devops 库中 zt_repo 表的 id 字段*****

// BranchType 分支类型模型
type BranchType struct {
	ID          uint            `gorm:"column:id"`
	Repo        uint            `gorm:"column:repo"`
	Name        string          `gorm:"column:name"`
	Key         string          `gorm:"column:key"`
	Prefixes    string          `gorm:"column:prefix"`
	Desc        string          `gorm:"column:desc"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     uint8           `gorm:"column:deleted"`
}

// TableName 显式指定结构体对应的数据表名为 ops_branch_type
func (BranchType) TableName() string {
	return "ops_branch_type"
}

// BeforeCreate GORM钩子：创建前自动填充创建时间与更新时间
func (b *BranchType) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	b.CreatedDate = now
	b.EditedDate = now
	return nil
}

// BeforeUpdate GORM钩子：更新前自动填充更新时间
func (b *BranchType) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	b.EditedDate = now
	return nil
}
