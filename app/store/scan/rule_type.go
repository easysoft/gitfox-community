// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package scan

import (
	"database/sql"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// Rule 扫描规则表
type Rule struct {
	ID          uint64                    `gorm:"column:id;primaryKey"`
	RuleKey     string                    `gorm:"column:ruleKey"`
	Name        string                    `gorm:"column:name"`
	Desc        string                    `gorm:"column:desc"`
	DescEn      string                    `gorm:"column:descEn"`
	Lang        constant.ScanLangName     `gorm:"column:lang"`
	Plugin      string                    `gorm:"column:plugin"`
	Tag         string                    `gorm:"column:tag"`
	Priority    constant.ScanRulePriority `gorm:"column:priority"`
	Type        constant.ScanRuleType     `gorm:"column:type"`
	Content     sql.NullString            `gorm:"column:content"`
	ContentEn   sql.NullString            `gorm:"column:contentEn"`
	IsCustom    uint8                     `gorm:"column:isCustom"`
	Status      constant.ScanRuleStatus   `gorm:"column:status"`
	CreatedBy   string                    `gorm:"column:createdBy"`
	CreatedDate *datetime.Local           `gorm:"column:createdDate"`
	EditedBy    string                    `gorm:"column:editedBy"`
	EditedDate  *datetime.Local           `gorm:"column:editedDate"`
	Deleted     uint8                     `gorm:"column:deleted"`
}

// TableName 指定数据库表名
func (r *Rule) TableName() string {
	return "ops_scan_rules"
}

// BeforeCreate GORM 钩子：创建前自动设置时间
func (r *Rule) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *Rule) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}
