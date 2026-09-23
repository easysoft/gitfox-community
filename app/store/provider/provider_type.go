// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// Provider 外部服务表（ops_provider）Store层结构体
type Provider struct {
	ID          uint            `gorm:"column:id;comment:'外部服务ID'"`
	Type        string          `gorm:"column:type;comment:'外部服务类型'"`
	Name        string          `gorm:"column:name;comment:'外部服务名称'"`
	URL         string          `gorm:"column:url;comment:'外部服务地址'"`
	Token       string          `gorm:"column:token;comment:'外部服务调用Token'"`
	CreatedBy   string          `gorm:"column:createdBy;comment:'创建人'"`
	CreatedDate *datetime.Local `gorm:"column:createdDate;comment:'创建时间'"`
	EditedBy    string          `gorm:"column:editedBy;comment:'更新人'"`
	EditedDate  *datetime.Local `gorm:"column:editedDate;comment:'更新时间'"`
	Deleted     uint8           `gorm:"column:deleted;comment:'删除标识（0=未删除，1=已删除）'"`
}

// TableName 指定数据库表名。
func (p *Provider) TableName() string {
	return "ops_provider"
}

// BeforeCreate GORM 钩子：创建前自动设置时间。
func (p *Provider) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.CreatedDate = now
	p.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate。
func (p *Provider) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.EditedDate = now
	return nil
}
