package scan

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gorm.io/gorm"
)

// RuleSetRules 规则集与规则关联表
type RuleSetRules struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	RuleSetID   int64      `gorm:"column:rulesetID"`
	RuleID      int64      `gorm:"column:ruleID"`
	CreatedBy   string     `gorm:"column:createdBy"`
	CreatedDate *time.Time `gorm:"column:createdDate"`
	EditedBy    string     `gorm:"column:editedBy"`
	EditedDate  *time.Time `gorm:"column:editedDate"`
	Deleted     bool       `gorm:"column:deleted"`
}

// TableName 指定数据库表名
func (r *RuleSetRules) TableName() string {
	return "ops_scan_ruleset_rules"
}

func (r *RuleSetRules) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	r.CreatedDate = &now
	r.EditedDate = &now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *RuleSetRules) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	r.EditedDate = &now
	return nil
}

// SolutionRules 扫描方案
type SolutionRules struct {
	ID          int64                   `gorm:"column:id;primaryKey"`
	Name        string                  `gorm:"column:name"`
	Desc        string                  `gorm:"column:desc"`
	DescEn      string                  `gorm:"column:descEn"`
	Tag         string                  `gorm:"column:tag"`
	IsCustom    bool                    `gorm:"column:isCustom"`
	Status      constant.ScanRuleStatus `gorm:"column:status"`
	CreatedBy   string                  `gorm:"column:createdBy"`
	CreatedDate *time.Time              `gorm:"column:createdDate"`
	EditedBy    string                  `gorm:"column:editedBy"`
	EditedDate  *time.Time              `gorm:"column:editedDate"`
	Deleted     bool                    `gorm:"column:deleted"`
}

// TableName 指定数据库表名
func (s *SolutionRules) TableName() string {
	return "ops_scan_solutions"
}

func (r *SolutionRules) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	r.CreatedDate = &now
	r.EditedDate = &now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *SolutionRules) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	r.EditedDate = &now
	return nil
}

// SolutionRuleSets 扫描方案与规则集关联表
type SolutionRuleSets struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	SolutionID  int64      `gorm:"column:solutionID"`
	RuleSetID   int64      `gorm:"column:ruleSetID"`
	CreatedBy   string     `gorm:"column:createdBy"`
	CreatedDate *time.Time `gorm:"column:createdDate"`
	EditedBy    string     `gorm:"column:editedBy"`
	EditedDate  *time.Time `gorm:"column:editedDate"`
	Deleted     bool       `gorm:"column:deleted"`
}

// TableName 指定数据库表名
func (s *SolutionRuleSets) TableName() string {
	return "ops_scan_solution_rulesets"
}

func (r *SolutionRuleSets) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	r.CreatedDate = &now
	r.EditedDate = &now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *SolutionRuleSets) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	r.EditedDate = &now
	return nil
}

type RuleMigration struct {
	Name        string     `gorm:"column:name"`
	Version     string     `gorm:"column:version"`
	CreatedBy   string     `gorm:"column:createdBy"`
	CreatedDate *time.Time `gorm:"column:createdDate"`
	Deleted     bool       `gorm:"column:deleted"`
}

func (r *RuleMigration) TableName() string {
	return "ops_scan_rule_migrations"
}
