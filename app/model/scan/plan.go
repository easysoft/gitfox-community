package scan

import (
	"time"
)

// ScanPlanReqForCreate 创建扫描计划请求
type ScanPlanReqForCreate struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Desc        string  `json:"desc"`
	SolutionIDs []uint  `json:"solutionIDs"`
	ScanType    string  `json:"scanType"`
	Branches    string  `json:"branches"`
	Files       string  `json:"files"`
}

// ScanPlanReqForUpdate 更新扫描计划请求
type ScanPlanReqForUpdate struct {
	Name     string `json:"name,omitempty"`
	Desc     string `json:"desc,omitempty"`
	ScanType string `json:"scanType,omitempty"`
	Branches string `json:"branches,omitempty"`
	Files    string `json:"files,omitempty"`
}

// ScanPlanReqFilter 查询扫描计划过滤条件
type ScanPlanReqFilter struct {
	Name     string `json:"name"`
	Sort     string `json:"sort" form:"sort" default:"id"`
	Order    string `json:"order" form:"order" default:"desc"`
	Page     int    `json:"page" form:"page" default:"1"`
	PageSize int    `json:"pageSize" form:"pageSize" default:"30"`
}

// ScanPlanReqForBind 绑定/解绑扫描方案请求
type ScanPlanReqForBind struct {
	SolutionIDs []uint `json:"solutionIDs"`
}

// ScanPlan 扫描计划响应模型
type ScanPlan struct {
	ID            int64      `json:"id"`
	RepoID        int64      `json:"repoID"`
	Name          string     `json:"name"`
	Desc          string     `json:"desc"`
	ScanType      string     `json:"scanType"`
	Branches      string     `json:"branches"`
	Files         string     `json:"files"`
	SolutionIDs   []int64    `json:"solutionIDs,omitempty"`
	SolutionCount int        `json:"solutionCount"`
	TriggerCount  int        `json:"triggerCount"`
	Solutions     []*Solution `json:"solutions,omitempty"`
	CreatedBy     string     `json:"createdBy,omitempty"`
	CreatedDate   *time.Time `json:"createdDate"`
	EditedBy      string     `json:"editedBy"`
	EditedDate    *time.Time `json:"editedDate"`
	Deleted       bool       `json:"deleted"`
}
