package jobs

import (
	"database/sql"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// ScheduleJob 后台任务模型
type ScheduleJob struct {
	ID                  int64           `gorm:"column:id;primaryKey;autoIncrement;comment:'任务ID，自增主键'"`
	Name                string          `gorm:"column:name"`
	Type                string          `gorm:"column:type"`
	Priority            int             `gorm:"column:priority"`
	Data                sql.NullString  `gorm:"column:data"`
	Result              sql.NullString  `gorm:"column:result"`
	MaxDurationSeconds  int             `gorm:"column:maxDurationSeconds"`
	MaxRetries          int             `gorm:"column:maxRetries"`
	State               string          `gorm:"column:state"`
	Scheduled           *datetime.Local `gorm:"column:scheduled"`
	TotalExecutions     int             `gorm:"column:totalExecutions"`
	RunBy               string          `gorm:"column:runBy"`
	RunDeadline         *datetime.Local `gorm:"column:runDeadline"`
	RunProgress         int             `gorm:"column:runProgress"`
	LastExecuted        *datetime.Local `gorm:"column:lastExecuted"`
	IsRecurring         bool            `gorm:"column:isRecurring"`
	RecurringCron       string          `gorm:"column:recurringCron"`
	ConsecutiveFailures int             `gorm:"column:consecutiveFailures"`
	LastFailureError    sql.NullString  `gorm:"column:lastFailureError"`
	GroupID             string          `gorm:"column:groupID"`
	CreatedBy           string          `gorm:"column:createdBy"`
	CreatedDate         *datetime.Local `gorm:"column:createdDate"`
	EditedBy            string          `gorm:"column:editedBy"`
	EditedDate          *datetime.Local `gorm:"column:editedDate"`
	Deleted             uint8           `gorm:"column:deleted"`
}

// TableName 指定数据库表名
func (ScheduleJob) TableName() string {
	return "ops_schedule_jobs"
}

// BeforeCreate GORM 钩子：创建前自动设置时间
func (j *ScheduleJob) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	j.CreatedDate = now
	j.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (j *ScheduleJob) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	j.EditedDate = now
	return nil
}
