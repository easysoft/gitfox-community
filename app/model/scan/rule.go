package scan

import "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

type Rule struct {
	ID            int64                     `json:"id,omitempty"`
	Name          string                    `json:"name"`
	Description   string                    `json:"description"`
	DescriptionEN string                    `json:"descriptionEN,omitempty"`
	Lang          constant.ScanLangName     `json:"lang"`
	LangID        constant.ScanLangID       `json:"langID,omitempty"`
	Plugin        string                    `json:"plugin"`
	Tag           string                    `json:"tag,omitempty"`
	Priority      constant.ScanRulePriority `json:"priority"`
	Status        constant.ScanRuleStatus   `json:"status,omitempty"`
	Type          constant.ScanRuleType     `json:"type"`
	Content       string                    `json:"content,omitempty"`
	ContentEN     string                    `json:"contentEN,omitempty"`
	CreatedBy     string                    `json:"createdBy,omitempty"`
	UpdatedBy     string                    `json:"updatedBy,omitempty"`
	CreatedAt     int64                     `json:"createdAt,omitempty"`
	UpdatedAt     int64                     `json:"updatedAt,omitempty"`
	IsCustom      bool                      `json:"isCustom,omitempty"`
	RuleKey       string                    `json:"ruleKey"`
}
