package scan

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
)

type RuleSetCreateReq struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	DescEn string `json:"descEn,omitempty"`
	Lang   string `json:"lang"`
	Tag    string `json:"tag"`
	Type   string `json:"type"`
	Plugin string `json:"plugin"`
}

type RuleSetUpdateReq struct {
	Name   string `json:"name,omitempty"`
	Desc   string `json:"desc,omitempty"`
	DescEn string `json:"descEn,omitempty"`
	Lang   string `json:"lang,omitempty"`
	Tag    string `json:"tag,omitempty"`
	Type   string `json:"type,omitempty"`
	Plugin string `json:"plugin,omitempty"`
	Status string `json:"status,omitempty"`
}

type RuleSet struct {
	ID          int64                   `json:"id"`
	Name        string                  `json:"name"`
	Desc        string                  `json:"desc"`
	DescEn      string                  `json:"descEn,omitempty"`
	Lang        constant.ScanLangName   `json:"lang"`
	Tag         string                  `json:"tag"`
	IsCustom    bool                    `json:"isCustom"`
	Status      constant.ScanRuleStatus `json:"status"`
	CreatedBy   string                  `json:"createdBy,omitempty"`
	CreatedDate *time.Time              `json:"createdDate"` // 允许 NULL，用指针类型
	EditedBy    string                  `json:"editedBy"`
	EditedDate  *time.Time              `json:"editedDate"` // 允许 NULL，用指针类型
	Deleted     bool                    `json:"deleted"`
	RulesCount  int64                   `json:"rulesCount"`
	Rules       []*Rule                 `json:"rules,omitempty"`
}
