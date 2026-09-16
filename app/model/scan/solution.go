package scan

import (
	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

// SolutionReqForCreate 创建扫描方案请求
type SolutionReqForCreate struct {
	Name   string                  `json:"name" binding:"required,max=100"`
	Desc   string                  `json:"desc"`
	DescEn string                  `json:"descEn,omitempty"`
	Tag    string                  `json:"tag"`
	Status constant.ScanRuleStatus `json:"status,omitempty"`
}

type SolutionReqForUpdate struct {
	Name   *string                 `json:"name,omitempty"`
	Desc   *string                 `json:"desc"`   // 不带 omitempty：允许前端传空字符串清空描述
	DescEn *string                 `json:"descEn"` // 不带 omitempty：允许前端传空字符串清空英文描述
	Tag    *string                 `json:"tag,omitempty"`
	Status constant.ScanRuleStatus `json:"status,omitempty"`
}

// SolutionReqFilter 查询扫描方案过滤条件
type SolutionReqFilter struct {
	ID        uint   `json:"id,omitempty"`
	Name      string `json:"name"`
	Tag       string `json:"tag"`
	Status    string `json:"status"`
	CreatedBy string `json:"createdBy"`
	EditedBy  string `json:"editedBy"`
	Sort      string `json:"sort" form:"sort" default:"id"`
	Order     string `json:"order" form:"order" default:"asc"`
	Page      int    `json:"page" form:"page" default:"1"`
	PageSize  int    `json:"pageSize" form:"pageSize" default:"30"`
}

// SolutionReqForBind 绑定/解绑规则集请求
type SolutionReqForBind struct {
	IDs []uint `json:"rulesetIDList"`
}

// Solution 扫描方案响应模型
type Solution struct {
	ID            int64                   `json:"id"`
	Name          string                  `json:"name"`
	Desc          string                  `json:"desc"`
	DescEn        string                  `json:"descEn,omitempty"`
	Tag           string                  `json:"tag"`
	IsCustom      bool                    `json:"isCustom"`
	Status        constant.ScanRuleStatus `json:"status"`
	CreatedBy     string                  `json:"createdBy,omitempty"`
	CreatedDate   *datetime.Local         `json:"createdDate"`
	EditedBy      string                  `json:"editedBy"`
	EditedDate    *datetime.Local         `json:"editedDate"`
	Deleted       bool                    `json:"deleted"`
	RuleSets      []*RuleSet              `json:"ruleSets,omitempty"`
	RuleSetsCount int64                   `json:"ruleSetsCount"`
	RulesCount    int64                   `json:"rulesCount"`
	Plugins       []string                `json:"plugins,omitempty"`
	Langs         []string                `json:"langs,omitempty"`
}
