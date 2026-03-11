package scan

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
)

type RuleSetReqForCreate struct {
	Name     string `json:"name" binding:"required,max=100"`
	Desc     string `json:"desc"`
	DescEn   string `json:"descEn,omitempty"`
	Lang     string `json:"lang"`
	Tag      string `json:"tag"`
	Type     string `json:"type" binding:"required,oneof=defect security compliance optimize"`
	Plugin   string `json:"plugin"`
	IsCustom bool   `json:"isCustom"`
}

type RuleSetReqForUpdate struct {
	Name   string `json:"name,omitempty"`
	Desc   string `json:"desc,omitempty"`
	DescEn string `json:"descEn,omitempty"`
	Lang   string `json:"lang,omitempty"`
	Tag    string `json:"tag,omitempty"`
	Type   string `json:"type,omitempty" binding:"omitempty,oneof=defect security compliance optimize"`
	Plugin string `json:"plugin,omitempty"`
	Status string `json:"status,omitempty"`
}

type RuleSetReqFilter struct {
	Name     string `json:"name"`
	Lang     string `json:"lang"`
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
	CreatedDate *time.Time              `json:"createdDate"` // 允许 NULL，用指针类型
	EditedBy    string                  `json:"editedBy"`
	EditedDate  *time.Time              `json:"editedDate"` // 允许 NULL，用指针类型
	Deleted     bool                    `json:"deleted"`
	RulesCount  int64                   `json:"rulesCount"`
}

type RuleSetReqForBind struct {
	IDs []uint `json:"ruleIDList"`
}
