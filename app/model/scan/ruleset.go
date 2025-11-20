package scan

import "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

type RuleSetCreateReq struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	DescriptionEN string `json:"descriptionEN,omitempty"`
	Lang          string `json:"lang"`
	Tag           string `json:"tag"`
	Type          string `json:"type"`
	Plugin        string `json:"plugin"`
}

type RuleSetUpdateReq struct {
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	DescriptionEN string `json:"descriptionEN,omitempty"`
	Lang          string `json:"lang,omitempty"`
	Tag           string `json:"tag,omitempty"`
	Type          string `json:"type,omitempty"`
	Plugin        string `json:"plugin,omitempty"`
	Status        string `json:"status,omitempty"`
}

type RuleSet struct {
	ID            int64                   `json:"id"`
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	DescriptionEN string                  `json:"descriptionEN,omitempty"`
	Lang          constant.ScanLangName   `json:"lang"`
	LangID        constant.ScanLangID     `json:"langID"`
	Tag           string                  `json:"tag"`
	Type          constant.ScanRuleType   `json:"type"`
	Plugin        string                  `json:"plugin,omitempty"`
	Status        constant.ScanRuleStatus `json:"status"`
	CreatedBy     string                  `json:"createdBy"`
	UpdatedBy     string                  `json:"updatedBy"`
	CreatedAt     int64                   `json:"createdAt"`
	UpdatedAt     int64                   `json:"updatedAt"`
	RulesCount    int64                   `json:"rulesCount"`
	Rules         []*Rule                 `json:"rules,omitempty"`
	IsCustom      bool                    `json:"isCustom"`
}
