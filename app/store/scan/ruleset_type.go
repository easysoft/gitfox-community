// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package scan

import (
	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

type RuleSet struct {
	ID          uint64                  `gorm:"column:id"`
	Name        string                  `gorm:"column:name"`
	Desc        string                  `gorm:"column:desc"`
	DescEn      string                  `gorm:"column:descEn"`
	Type        constant.ScanRuleType   `gorm:"column:type"`
	Lang        constant.ScanLangName   `gorm:"column:lang"`
	Plugin      string                  `gorm:"column:plugin"`
	Tag         string                  `gorm:"column:tag"`
	IsCustom    uint8                   `gorm:"column:isCustom"`
	Status      constant.ScanRuleStatus `gorm:"column:status"`
	CreatedBy   string                  `gorm:"column:createdBy"`
	CreatedDate *datetime.Local         `gorm:"column:createdDate"`
	EditedBy    string                  `gorm:"column:editedBy"`
	EditedDate  *datetime.Local         `gorm:"column:editedDate"`
	Deleted     uint8                   `gorm:"column:deleted;comment:'删除标识（0=未删除，1=已删除）'"`
}

// TableName 指定数据库表名（GORM自动映射时使用）
func (r *RuleSet) TableName() string {
	return "ops_scan_rulesets"
}

func (r *RuleSet) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *RuleSet) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}
