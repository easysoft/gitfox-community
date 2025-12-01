package scan

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
)

type Rule struct {
	ID          int64                     `json:"id,omitempty"`
	RuleKey     string                    `json:"ruleKey"`
	Name        string                    `json:"name"`
	Desc        string                    `json:"desc"`
	DescEn      string                    `json:"descEn,omitempty"`
	Lang        constant.ScanLangName     `json:"lang"`
	Plugin      string                    `json:"plugin"`
	Tag         string                    `json:"tag,omitempty"`
	Priority    constant.ScanRulePriority `json:"priority"`
	Type        constant.ScanRuleType     `json:"type"`
	Content     string                    `json:"content,omitempty"`
	ContentEn   string                    `json:"contentEn,omitempty"`
	IsCustom    bool                      `json:"isCustom,omitempty"`
	Status      constant.ScanRuleStatus   `json:"status,omitempty"`
	CreatedBy   string                    `json:"createdBy,omitempty"`
	CreatedDate *time.Time                `json:"createdDate"` // 允许 NULL，用指针类型
	EditedBy    string                    `json:"editedBy"`
	EditedDate  *time.Time                `json:"editedDate"` // 允许 NULL，用指针类型
	Deleted     bool                      `json:"deleted"`
}

type RuleReqFilter struct {
	Sort  string `json:"sort" form:"sort" default:"id"`
	Order string `json:"order" form:"order" default:"asc"`
	Page  int    `json:"page" form:"page" default:"1"`
	Size  int    `json:"size" form:"size" default:"30"`
}
