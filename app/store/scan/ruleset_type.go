package scan

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gorm.io/gorm"
)

type RuleSet struct {
	ID            int64                   `gorm:"column:id"`
	Name          string                  `gorm:"column:name"`
	Description   string                  `gorm:"column:description"`
	DescriptionEN string                  `gorm:"column:descriptionEN"`
	Lang          constant.ScanLangName   `gorm:"column:lang"`
	LangID        constant.ScanLangID     `gorm:"column:langID"`
	Plugin        string                  `gorm:"column:plugin"`
	Tag           string                  `gorm:"column:tag"`
	Type          constant.ScanRuleType   `gorm:"column:type"`
	Status        constant.ScanRuleStatus `gorm:"column:status"`
	CreatedBy     string                  `gorm:"column:createdBy"`
	UpdatedBy     string                  `gorm:"column:updatedBy"`
	CreatedDate   *time.Time              `gorm:"column:createdDate"`
	EditedDate    *time.Time              `gorm:"column:updatedDate"`
	Deleted       bool                    `gorm:"column:deleted;comment:'删除标识（0=未删除，1=已删除）'"`
	IsCustom      bool                    `gorm:"column:isCustom"`
}

// TableName 指定数据库表名（GORM自动映射时使用）
func (r *RuleSet) TableName() string {
	return "ops_ruleset"
}

func (r *RuleSet) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	r.CreatedDate = &now
	r.EditedDate = &now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *RuleSet) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	r.EditedDate = &now
	return nil
}
