package scan

import (
	"time"
)

// ScanConditionReqForBatchCreate 批量创建条件请求（trigger-scoped）
type ScanConditionReqForBatchCreate struct {
	Conditions []ScanConditionItem `json:"conditions" binding:"required"`
}

// ScanConditionItem 单个条件项
type ScanConditionItem struct {
	Priority  ScanRulePriority `json:"priority"`
	Type      ScanRuleType     `json:"type"`
	Unit      ScanRuleUnit     `json:"unit"`
	Threshold int64            `json:"threshold"`
}

// ScanConditionReqForBatchDelete 批量删除条件请求（trigger-scoped）
type ScanConditionReqForBatchDelete struct {
	IDs []uint `json:"ids" binding:"required"`
}

// ScanPlanConditionReqForCreate 创建计划级条件请求（plan-scoped, triggerID=0）
type ScanPlanConditionReqForCreate struct {
	Priority  ScanRulePriority `json:"priority"`
	Type      ScanRuleType     `json:"type"`
	Unit      ScanRuleUnit     `json:"unit"`
	Threshold int64            `json:"threshold"`
}

// ScanPlanConditionReqForUpdate 更新计划级条件请求（plan-scoped, PATCH）
//
// Pointer fields are used to distinguish "not provided" from "zero".
type ScanPlanConditionReqForUpdate struct {
	Priority  *ScanRulePriority `json:"priority"`
	Type      *ScanRuleType     `json:"type"`
	Unit      *ScanRuleUnit     `json:"unit"`
	Threshold *int64            `json:"threshold"`
}

// ScanPlanConditionReqForBatchCreate 批量创建计划级条件请求（plan-scoped）
type ScanPlanConditionReqForBatchCreate struct {
	Conditions []ScanPlanConditionReqForCreate `json:"conditions" binding:"required"`
}

// ScanPlanConditionReqForBatchDelete 批量删除计划级条件请求（plan-scoped）
type ScanPlanConditionReqForBatchDelete struct {
	IDs []uint `json:"ids" binding:"required"`
}

// ScanCondition 触发器条件 / 计划级条件响应模型
type ScanCondition struct {
	ID          int64            `json:"id"`
	PlanID      int64            `json:"planID"`
	TriggerID   int64            `json:"triggerID"`
	RepoID      int64            `json:"repoID"`
	Priority    ScanRulePriority `json:"priority"`
	Type        ScanRuleType     `json:"type"`
	Unit        ScanRuleUnit     `json:"unit"`
	Threshold   int64            `json:"threshold"`
	CreatedBy   string           `json:"createdBy,omitempty"`
	CreatedDate *time.Time       `json:"createdDate"`
	EditedBy    string           `json:"editedBy"`
	EditedDate  *time.Time       `json:"editedDate"`
	Deleted     bool             `json:"deleted"`
}
