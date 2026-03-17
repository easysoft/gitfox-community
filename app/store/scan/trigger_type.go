package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// ScanPlanTrigger 扫描计划触发器 DB 模型
type ScanPlanTrigger struct {
	ID          uint64          `gorm:"column:id;primaryKey"`
	PlanID      uint64          `gorm:"column:planID"`
	RepoID      uint64          `gorm:"column:repoID"`
	Name        string          `gorm:"column:name"`
	Desc        string          `gorm:"column:desc"`
	Type        string          `gorm:"column:type"`
	Cron        string          `gorm:"column:cron"`
	CronBranch  string          `gorm:"column:cronBranch"`
	Keywords    string          `gorm:"column:keywords"`
	Actions     string          `gorm:"column:actions"`
	Disabled    bool            `gorm:"column:disabled"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     bool            `gorm:"column:deleted"`
}

func (r *ScanPlanTrigger) TableName() string { return "ops_scan_triggers" }

func (r *ScanPlanTrigger) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

func (r *ScanPlanTrigger) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}
