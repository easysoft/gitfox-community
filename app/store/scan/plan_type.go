package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// ScanPlan 扫描计划 DB 模型
type ScanPlan struct {
	ID          uint64          `gorm:"column:id;primaryKey"`
	RepoID      uint64          `gorm:"column:repoID"`
	Name        string          `gorm:"column:name"`
	Desc        string          `gorm:"column:desc"`
	ScanType    string          `gorm:"column:scanType"`
	Branches    string          `gorm:"column:branches"`
	Files       string          `gorm:"column:files"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     uint8           `gorm:"column:deleted"`
}

func (r *ScanPlan) TableName() string { return "ops_scan_plans" }

func (r *ScanPlan) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

func (r *ScanPlan) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}

// ScanPlanSolution 扫描计划与方案关联表
type ScanPlanSolution struct {
	ID          uint64          `gorm:"column:id;primaryKey"`
	PlanID      uint64          `gorm:"column:planID"`
	SolutionID  uint64          `gorm:"column:solutionID"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     uint8           `gorm:"column:deleted"`
}

func (r *ScanPlanSolution) TableName() string { return "ops_scan_plan_solutions" }

func (r *ScanPlanSolution) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

func (r *ScanPlanSolution) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}
