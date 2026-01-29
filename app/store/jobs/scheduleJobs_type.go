package jobs

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// ScheduleJob 后台任务模型
type ScheduleJob struct {
	ID                  int64           `gorm:"column:id;primaryKey;autoIncrement;comment:'任务ID，自增主键'"`
	Name                string          `gorm:"column:name;not null;uniqueIndex:uk_schedule_job_name;comment:'任务唯一标识'"`
	Type                string          `gorm:"column:type;not null;index:idx_schedule_job_type;comment:'任务类型'"`
	Priority            int             `gorm:"column:priority;not null;default:0;comment:'任务优先级'"`
	Data                string          `gorm:"column:data;type:text;not null;comment:'任务数据（JSON格式）'"`
	Result              string          `gorm:"column:result;type:text;not null;comment:'任务结果（JSON格式）'"`
	MaxDurationSeconds  int             `gorm:"column:maxDurationSeconds;not null;default:0;comment:'最大执行时长（秒）'"`
	MaxRetries          int             `gorm:"column:maxRetries;not null;default:0;comment:'最大重试次数'"`
	State               string          `gorm:"column:state;not null;index:idx_schedule_job_state;comment:'任务状态'"`
	Scheduled           *datetime.Local `gorm:"column:scheduled;type:datetime;comment:'计划执行时间'"`
	TotalExecutions     int             `gorm:"column:totalExecutions;not null;default:0;comment:'总执行次数'"`
	RunBy               string          `gorm:"column:runBy;not null;default:'';comment:'执行者'"`
	RunDeadline         *datetime.Local `gorm:"column:runDeadline;type:datetime;comment:'执行截止时间'"`
	RunProgress         int             `gorm:"column:runProgress;not null;default:0;comment:'执行进度'"`
	LastExecuted        *datetime.Local `gorm:"column:lastExecuted;type:datetime;comment:'最后执行时间'"`
	IsRecurring         bool            `gorm:"column:isRecurring;not null;default:0;comment:'是否为周期性任务，0-否，1-是'"`
	RecurringCron       string          `gorm:"column:recurringCron;not null;default:'';comment:'周期性任务Cron表达式'"`
	ConsecutiveFailures int             `gorm:"column:consecutiveFailures;not null;default:0;comment:'连续失败次数'"`
	LastFailureError    string          `gorm:"column:lastFailureError;type:text;not null;comment:'最后失败错误信息'"`
	GroupID             string          `gorm:"column:groupID;not null;default:'';index:idx_schedule_job_groupID;comment:'任务组ID'"`
	CreatedBy           string          `gorm:"column:createdBy;not null;default:'';comment:'创建人'"`
	CreatedDate         *datetime.Local `gorm:"column:createdDate;type:datetime;comment:'创建时间'"`
	EditedBy            string          `gorm:"column:editedBy;not null;default:'';comment:'更新人'"`
	EditedDate          *datetime.Local `gorm:"column:editedDate;type:datetime;comment:'更新时间'"`
	Deleted             uint8           `gorm:"column:deleted;not null;default:0;comment:'删除标识，0-未删除，1-已删除'"`
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
