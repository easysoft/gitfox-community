// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package space

import (
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// 定义 Acl 字段的可选值常量（规范业务使用）
const (
	AclPrivate = "private" // 私有权限
	AclOpen    = "open"    // 公开权限
)

// Space 空间模型
type Space struct {
	ID          uint            `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement;comment:'ID'" json:"id"`
	Name        string          `gorm:"column:name;type:varchar(200);not null;default:'';comment:'名称'" json:"name"`
	Code        string          `gorm:"column:code;type:varchar(50);not null;uniqueIndex:uk_code;comment:'唯一标识'" json:"code"` // 唯一索引
	Auth        string          `gorm:"column:auth;type:varchar(30);not null;default:'';comment:'空间人员权限类型'" json:"auth"`
	ACL         string          `gorm:"column:acl;type:varchar(30);not null;default:'open';comment:'权限:private,open'" json:"acl"`
	Desc        *string         `gorm:"column:desc;type:varchar(500);not null;default:'';comment:'描述'" json:"desc,omitempty"`
	CreatedBy   string          `gorm:"column:createdBy;type:varchar(30);not null;default:'';comment:'创建人'" json:"createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate;type:datetime;comment:'创建时间'" json:"createdDate,omitempty"` // 允许为NULL，指针类型
	EditedBy    string          `gorm:"column:editedBy;type:varchar(30);not null;default:'';comment:'更新人'" json:"editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate;type:datetime;comment:'更新时间'" json:"editedDate,omitempty"` // 允许为NULL，指针类型
	Deleted     uint8           `gorm:"column:deleted;type:tinyint;not null;default:0;comment:'删除标识，0-未删除，1-已删除'" json:"deleted"`
}

// TableName 显式指定结构体对应的数据表名为 ops_space
func (Space) TableName() string {
	return "ops_space"
}

// BeforeCreate GORM钩子：创建前自动填充创建时间与更新时间
func (s *Space) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	s.CreatedDate = now
	s.EditedDate = now
	return nil
}

// BeforeUpdate GORM钩子：更新前自动填充更新时间
func (s *Space) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	s.EditedDate = now
	return nil
}
