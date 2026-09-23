// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package scan

import (
	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
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
