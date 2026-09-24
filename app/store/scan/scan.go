// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package scan

import (
	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// RuleSetRules 规则集与规则关联表
type RuleSetRules struct {
	ID          uint64          `gorm:"column:id;primaryKey"`
	RuleSetID   uint64          `gorm:"column:rulesetID"`
	RuleID      uint64          `gorm:"column:ruleID"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     uint8           `gorm:"column:deleted"`
}

// TableName 指定数据库表名
func (r *RuleSetRules) TableName() string {
	return "ops_scan_ruleset_rules"
}

func (r *RuleSetRules) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *RuleSetRules) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}

// SolutionRules 扫描方案
type SolutionRules struct {
	ID          uint64                  `gorm:"column:id;primaryKey"`
	Name        string                  `gorm:"column:name"`
	Desc        string                  `gorm:"column:desc"`
	DescEn      string                  `gorm:"column:descEn"`
	Tag         string                  `gorm:"column:tag"`
	IsCustom    uint8                   `gorm:"column:isCustom"`
	Status      constant.ScanRuleStatus `gorm:"column:status"`
	CreatedBy   string                  `gorm:"column:createdBy"`
	CreatedDate *datetime.Local         `gorm:"column:createdDate"`
	EditedBy    string                  `gorm:"column:editedBy"`
	EditedDate  *datetime.Local         `gorm:"column:editedDate"`
	Deleted     uint8                   `gorm:"column:deleted"`
}

// TableName 指定数据库表名
func (s *SolutionRules) TableName() string {
	return "ops_scan_solutions"
}

func (r *SolutionRules) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *SolutionRules) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}

// SolutionRuleSets 扫描方案与规则集关联表
type SolutionRuleSets struct {
	ID          uint64          `gorm:"column:id;primaryKey"`
	SolutionID  uint64          `gorm:"column:solutionID"`
	RuleSetID   uint64          `gorm:"column:rulesetID"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
	Deleted     uint8           `gorm:"column:deleted"`
}

// TableName 指定数据库表名
func (s *SolutionRuleSets) TableName() string {
	return "ops_scan_solution_rulesets"
}

func (r *SolutionRuleSets) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *SolutionRuleSets) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}

type RuleMigration struct {
	Name        string          `gorm:"column:name"`
	Version     string          `gorm:"column:version"`
	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	Deleted     uint8           `gorm:"column:deleted"`
}

func (r *RuleMigration) TableName() string {
	return "ops_scan_rule_migrations"
}
