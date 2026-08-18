package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// AIReviewRule 规则模型
type AIReviewRule struct {
	ID          uint            `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement;comment:'ID'" json:"id"`
	Language    string          `gorm:"column:language;type:varchar(50);not null;default:'';comment:'编程语言'" json:"language"`
	Extension   string          `gorm:"column:extension;type:varchar(100);not null;default:'';comment:'文件后缀名'" json:"extension"`
	Rule        string          `gorm:"column:rule;type:text;default:null;comment:'评审规则文本'" json:"rule"`
	CreatedBy   string          `gorm:"column:createdBy;type:varchar(30);not null;default:'';comment:'创建人'" json:"createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate;type:datetime;comment:'创建时间'" json:"createdDate,omitempty"` // 允许为NULL，指针类型
	EditedBy    string          `gorm:"column:editedBy;type:varchar(30);not null;default:'';comment:'更新人'" json:"editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate;type:datetime;comment:'更新时间'" json:"editedDate,omitempty"` // 允许为NULL，指针类型
	Deleted     uint8           `gorm:"column:deleted;type:tinyint;not null;default:0;comment:'删除标识，0-未删除，1-已删除'" json:"deleted"`
}

// TableName 显式指定结构体对应的数据表名为 ops_ai_review_rules
func (AIReviewRule) TableName() string {
	return "ops_ai_review_rules"
}

// BeforeCreate GORM钩子：创建前自动填充创建时间与更新时间
func (s *AIReviewRule) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	s.CreatedDate = now
	s.EditedDate = now
	return nil
}

// BeforeUpdate GORM钩子：更新前自动填充更新时间
func (s *AIReviewRule) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	s.EditedDate = now
	return nil
}
