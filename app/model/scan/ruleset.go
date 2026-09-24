// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package scan

import (
	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

type RuleSetReqForCreate struct {
	Name     string `json:"name" binding:"required,max=100"`
	Desc     string `json:"desc"`
	DescEn   string `json:"descEn,omitempty"`
	Lang     string `json:"lang"`
	Tag      string `json:"tag"`
	Type     string `json:"type"`
	Plugin   string `json:"plugin"`
	IsCustom bool   `json:"isCustom"`
}

type RuleSetReqForUpdate struct {
	Name   *string `json:"name,omitempty"`
	Desc   *string `json:"desc"`   // 不带 omitempty：允许前端传空字符串清空描述
	DescEn *string `json:"descEn"` // 不带 omitempty：允许前端传空字符串清空英文描述
	Lang   *string `json:"lang,omitempty"`
	Tag    *string `json:"tag,omitempty"`
	Type   string  `json:"type,omitempty"`
	Plugin string  `json:"plugin,omitempty"`
	Status string  `json:"status,omitempty"`
}

type RuleSetReqFilter struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Lang     string              `json:"lang"`
	LangID   constant.ScanLangID `json:"langID"`
	Plugin   string              `json:"plugin"`
	Tag      string `json:"tag"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Sort     string `json:"sort" form:"sort" default:"id"`
	Order    string `json:"order" form:"order" default:"asc"`
	Page     int    `json:"page" form:"page" default:"1"`
	PageSize int    `json:"pageSize" form:"pageSize" default:"30"`
}

type RuleSet struct {
	ID          int64                   `json:"id"`
	Name        string                  `json:"name"`
	Desc        string                  `json:"desc"`
	DescEn      string                  `json:"descEn,omitempty"`
	Lang        constant.ScanLangName   `json:"lang"`
	Plugin      string                  `json:"plugin,omitempty"`
	Tag         string                  `json:"tag"`
	IsCustom    bool                    `json:"isCustom"`
	Status      constant.ScanRuleStatus `json:"status"`
	CreatedBy   string                  `json:"createdBy,omitempty"`
	CreatedDate *datetime.Local         `json:"createdDate"` // 允许 NULL，用指针类型
	EditedBy    string                  `json:"editedBy"`
	EditedDate  *datetime.Local         `json:"editedDate"` // 允许 NULL，用指针类型
	Deleted     bool                    `json:"deleted"`
	RulesCount  int64                   `json:"rulesCount"`
}

type RuleSetReqForBind struct {
	IDs []uint `json:"ruleIDList"`
}
