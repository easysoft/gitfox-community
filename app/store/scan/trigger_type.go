// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package scan

import (
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
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
	ScanType    string          `gorm:"column:scanType"`
	Cron        string          `gorm:"column:cron"`
	CronBranch  string          `gorm:"column:cronBranch"`
	Keywords    string          `gorm:"column:keywords"`
	Actions     string          `gorm:"column:actions"`
	Disabled    uint8           `gorm:"column:disabled"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     uint8           `gorm:"column:deleted"`
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
