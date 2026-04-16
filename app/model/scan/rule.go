package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
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
	CreatedDate *datetime.Local           `json:"createdDate"` // 允许 NULL，用指针类型
	EditedBy    string                    `json:"editedBy"`
	EditedDate  *datetime.Local           `json:"editedDate"` // 允许 NULL，用指针类型
	Deleted     bool                      `json:"deleted"`
}

type RuleReqFilter struct {
	ID       uint                `json:"id,omitempty"`
	Name     string              `json:"name,omitempty"`
	Plugin   string              `json:"plugin,omitempty"`
	Priority string              `json:"priority,omitempty"`
	Lang     string              `json:"lang,omitempty"`
	LangID   constant.ScanLangID `json:"langID,omitempty"`
	Tag      string              `json:"tag,omitempty"`
	Status   string              `json:"status,omitempty"`
	Type     string              `json:"type,omitempty"`
	Sort     string              `json:"sort" default:"id"`
	Order    string              `json:"order" default:"asc"`
	Page     int                 `json:"page" default:"1"`
	Size     int                 `json:"pageSize" default:"30"`
}

type RuleListReqFilter struct {
	ID           uint                      `json:"id,omitempty"`
	RuleSetID    int64                     `json:"rulesetID,omitempty"`
	Name         string                    `json:"name,omitempty"`
	Lang         string                    `json:"lang,omitempty"`
	LangID       constant.ScanLangID       `json:"langID,omitempty"`
	Plugin       string                    `json:"plugin,omitempty"`
	Tag          string                    `json:"tag,omitempty"`
	Priority     constant.ScanRulePriority `json:"priority,omitempty"`
	Status       constant.ScanRuleStatus   `json:"status,omitempty"`
	Type         constant.ScanRuleType     `json:"type,omitempty"`
	Sort         constant.ScanRuleSort     `json:"sort"`
	Order        constant.Order            `json:"order"`
	Bind         constant.ScanRuleBind     `json:"bind,omitempty"`
	ShowAllRules bool                      `json:"showAllRules,omitempty"`
	Page         int                       `json:"page" form:"page"`
	PageSize     int                       `json:"pageSize" form:"pageSize"`
}

func (f *RuleListReqFilter) Validater() error {
	// 如果前端传了有效的 LangID (> 0)，则强制以 LangID 映射为准（优先级高于独立传入的字符串）
	if f.LangID > 0 {
		f.Lang = f.LangID.String()
	}

	// 设置分页默认值
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 {
		f.PageSize = 30
	}

	// 验证 Priority
	if f.Priority != "" && !f.Priority.IsValid() {
		return errors.Newf("priority %s is invalid", f.Priority)
	}

	// 验证 Status
	if f.Status != "" && !f.Status.IsValid() {
		return errors.Newf("status %s is invalid", f.Status)
	}

	// 验证 Type
	if f.Type != "" && !f.Type.IsValid() {
		return errors.Newf("type %s is invalid", f.Type)
	}

	// 验证和设置 Sort 默认值
	if f.Sort == "" {
		f.Sort = constant.ScanRuleSortID
	}
	if !f.Sort.IsValid() {
		return errors.Newf("sort %s is invalid", f.Sort)
	}

	// 验证和设置 Order 默认值
	if f.Order == "" {
		f.Order = constant.OrderDesc
	}
	if !f.Order.IsValid() {
		return errors.Newf("order %s is invalid", f.Order)
	}

	// 验证 Bind
	if f.Bind != "" && !f.Bind.IsValid() {
		return errors.Newf("bind %s is invalid", f.Bind)
	}

	return nil
}

type ScanBaseResp struct {
	ID     int64  `json:"id"`
	Tag    string `json:"tag,omitempty"`
	Lang   string `json:"lang,omitempty"`
	Plugin string `json:"plugin,omitempty"`
	Type   string `json:"type,omitempty"`
	Count  int64  `json:"count,omitempty"`
}
