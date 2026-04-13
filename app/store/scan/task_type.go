package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// ScanTask 扫描任务 DB 模型
type ScanTask struct {
	ID          uint64                       `gorm:"column:id;primaryKey"`
	PlanID      uint64                       `gorm:"column:planID"`
	ExecutionID uint64                       `gorm:"column:executionID"`
	RepoNumber  int64                        `gorm:"column:repoNumber"`
	PlanNumber  int64                        `gorm:"column:planNumber"`
	TriggerID   *uint64                      `gorm:"column:triggerID"`
	TriggerType constant.ScanPlanTriggerType `gorm:"column:triggerType"`
	Status      constant.ScanTaskStatus      `gorm:"column:status"`
	Result      *constant.ScanTaskResult     `gorm:"column:result"`
	CreatedBy   string                       `gorm:"column:createdBy"`
	CreatedDate *datetime.Local              `gorm:"column:createdDate"`
	EditedDate  *datetime.Local              `gorm:"column:editedDate"`
	ScanType    constant.ScanType            `gorm:"column:scanType"`
	IssueNumber int64                        `gorm:"column:issueNumber"`
	Deleted     bool                         `gorm:"column:deleted"`
}

func (r *ScanTask) TableName() string { return "ops_scan_tasks" }

// BeforeCreate GORM 钩子：创建前自动设置时间
func (r *ScanTask) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *ScanTask) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}

type taskJoin struct {
	ScanTask
	RepoId      int64  `gorm:"column:repoID"`
	RepoName    string `gorm:"column:repoName"`
	PlanName    string `gorm:"column:planName"`
	TriggerName string `gorm:"column:triggerName"`

	ExecutionNumber   int64             `gorm:"column:executionNumber"`
	ExecutionRef      string            `gorm:"column:executionRef"`
	ExecutionStatus   constant.CIStatus `gorm:"column:executionStatus"`
	ExecutionStarted  *datetime.Local   `gorm:"column:executionStarted"`
	ExecutionFinished *datetime.Local   `gorm:"column:executionFinished"`
}
