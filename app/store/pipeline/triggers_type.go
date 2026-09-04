package pipeline

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// Trigger 触发器表 Store 层结构体
type Trigger struct {
	ID          int64           `gorm:"column:id;primaryKey;autoIncrement;comment:'触发器ID，自增主键'"`
	RepoID      int64           `gorm:"column:repoID;not null;comment:'关联的仓库ID'"`
	PipelineID  int64           `gorm:"column:pipelineID;not null;comment:'关联的流水线ID'"`
	Comment     string          `gorm:"column:comment;not null;default:'';comment:'当事件为 branch_updated 时进行提交关键词匹配'"`
	Event       string          `gorm:"column:event;not null;default:'';comment:'事件触发动作'"`
	Cron        string          `gorm:"column:cron;not null;default:'';comment:'定时配置'"`
	CreatedBy   string          `gorm:"column:createdBy;not null;default:'';comment:'创建人'"`
	CreatedDate *datetime.Local `gorm:"column:createdDate;comment:'创建时间'"`
	EditedBy    string          `gorm:"column:editedBy;not null;default:'';comment:'更新人'"`
	EditedDate  *datetime.Local `gorm:"column:editedDate;comment:'更新时间'"`
	Deleted     uint8           `gorm:"column:deleted;not null;default:0;comment:'删除标识，0-未删除，1-已删除'"`
}

// TableName 指定数据库表名
func (p *Trigger) TableName() string {
	return "ops_triggers"
}

// BeforeCreate GORM 钩子：创建前自动设置时间
func (p *Trigger) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.CreatedDate = now
	p.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (p *Trigger) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.EditedDate = now
	return nil
}
