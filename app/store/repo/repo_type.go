// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package repo

import (
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// Repository 代码库表（ops_repo）Store层结构体
type Repository struct {
	ID            uint            `gorm:"column:id;comment:'代码库ID'"`
	SpaceID       uint            `gorm:"column:spaceID;comment:'空间ID'"`
	Identifier    string          `gorm:"column:name;comment:'代码库唯一标识'"`
	Desc          string          `gorm:"column:desc;comment:'代码库描述'"` // 关键字兼容：映射数据库 desc 字段
	SCMType       string          `gorm:"column:scmType;default:git;comment:'源代码管理类型'"`
	Product       string          `gorm:"column:product;comment:'关联产品'"`
	GitUID        string          `gorm:"column:gitUID;comment:'Git仓库唯一标识'"`
	DefaultBranch string          `gorm:"column:defaultBranch;comment:'默认分支名'"`
	ForkID        uint            `gorm:"column:forkID;comment:'派生来源代码库ID'"`
	ACL           string          `gorm:"column:acl;comment:'权限类型（private=私有，open=公开）'"`
	Mirror        uint8           `gorm:"column:mirror;comment:'是否为镜像库（0=普通库，1=镜像库）'"`
	ProviderID    uint            `gorm:"column:providerID;comment:'源代码库提供商ID'"`
	Connector     []byte          `gorm:"column:connector;comment:'源代码库连接器信息'"`
	Status        string          `gorm:"column:status;comment:'代码库状态（active,importing,archived）'"`
	Synced        uint8           `gorm:"column:synced;comment:'同步标识'"`
	SpaceCode     string          `gorm:"->;column:spaceCode;comment:'空间唯一标识(关联查询)'"`
	CreatedBy     string          `gorm:"column:createdBy;comment:'创建人'"`
	CreatedDate   *datetime.Local `gorm:"column:createdDate;comment:'创建时间'"`
	EditedBy      string          `gorm:"column:editedBy;comment:'更新人'"`
	EditedDate    *datetime.Local `gorm:"column:editedDate;comment:'更新时间'"`
	Deleted       uint8           `gorm:"column:deleted;comment:'删除标识（0=未删除，1=已删除）'"`
}

// TableName 指定数据库表名（GORM自动映射时使用）
func (r *Repository) TableName() string {
	return "ops_repo"
}

func (r *Repository) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *Repository) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}
