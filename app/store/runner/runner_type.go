package runner

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// Runner 数据库模型
type Runner struct {
	ID          uint            `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string          `gorm:"column:name;not null"`
	Desc        string          `gorm:"column:desc;not null;default:''"`
	Version     string          `gorm:"column:version;not null;default:''"`
	Hostname    string          `gorm:"column:hostname;not null;default:''"`
	IP          string          `gorm:"column:ip;not null;default:''"`
	OS          string          `gorm:"column:os;not null;default:''"`
	Arch        string          `gorm:"column:arch;not null;default:''"`
	Runtime     string          `gorm:"column:runtime;not null;default:''"`
	Labels      string          `gorm:"column:labels;not null;default:''"`
	Token       string          `gorm:"column:token;not null"`
	HeartBeat   int             `gorm:"column:heartBeat"`
	Online      string          `gorm:"column:online;not null;default:'offline'"`
	Status      string          `gorm:"column:status;not null;default:'disable'"`
	IsDefault   bool            `gorm:"column:isDefault;not null;default:false"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     uint8           `gorm:"column:deleted;comment:'删除标识（0=未删除，1=已删除）'"`
}

func (r *Runner) TableName() string {
	return "ops_runner"
}

// BeforeCreate GORM 钩子：创建前自动设置时间
func (r *Runner) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *Runner) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}
