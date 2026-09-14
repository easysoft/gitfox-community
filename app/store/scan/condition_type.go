package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// ScanPlanCondition 触发器条件 DB 模型
type ScanPlanCondition struct {
	ID          uint64                    `gorm:"column:id;primaryKey"`
	PlanID      uint64                    `gorm:"column:planID"`
	TriggerID   uint64                    `gorm:"column:triggerID"`
	RepoID      uint64                    `gorm:"column:repoID"`
	Priority    constant.ScanRulePriority `gorm:"column:priority"`
	Type        constant.ScanRuleType     `gorm:"column:type"`
	Unit        constant.ScanRuleUnit     `gorm:"column:unit"`
	Threshold   int64                     `gorm:"column:threshold"`
	CreatedBy   string                    `gorm:"column:createdBy"`
	CreatedDate *datetime.Local           `gorm:"column:createdDate"`
	EditedBy    string                    `gorm:"column:editedBy"`
	EditedDate  *datetime.Local           `gorm:"column:editedDate"`
	Deleted     uint8                     `gorm:"column:deleted"`
}

func (r *ScanPlanCondition) TableName() string { return "ops_scan_plan_conditions" }

func (r *ScanPlanCondition) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

func (r *ScanPlanCondition) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}
