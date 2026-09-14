package pipeline

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// Plugin 插件表 Store 层结构体
type Plugin struct {
	ID          int64           `gorm:"column:id;primaryKey;autoIncrement;comment:'插件ID，自增主键'"`
	GroupID     int64           `gorm:"column:groupID"`
	Name        string          `gorm:"column:name;uniqueIndex:uk_plugin_name"`
	Desc        []byte          `gorm:"column:desc"`
	Logo        []byte          `gorm:"column:logo"`
	Yaml        []byte          `gorm:"column:yaml"`
	Json        []byte          `gorm:"column:json"`
	Type        string          `gorm:"column:type"`
	Kind        string          `gorm:"column:kind"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     uint8           `gorm:"column:deleted"`
}

// TableName 指定数据库表名
func (p *Plugin) TableName() string {
	return "ops_plugins"
}

// BeforeCreate GORM 钩子：创建前自动设置时间
func (p *Plugin) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.CreatedDate = now
	p.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (p *Plugin) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.EditedDate = now
	return nil
}

type PluginGroup struct {
	ID      int64  `gorm:"column:id;primaryKey;autoIncrement;comment:'插件分组ID，自增主键'"`
	Name    string `gorm:"column:name;not null;comment:'插件分组名称（唯一标识符）'"`
	Desc    string `gorm:"column:desc;not null;default:'';comment:'插件分组描述信息'"`
	Deleted uint8  `gorm:"column:deleted;not null;default:0;comment:'删除标识，0-未删除，1-已删除'"`
}

// TableName 指定数据库表名
func (p *PluginGroup) TableName() string {
	return "ops_plugin_group"
}
