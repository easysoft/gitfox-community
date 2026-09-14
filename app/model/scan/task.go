package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

type ScanTask struct {
	ID          int64                        `json:"id"`
	PlanID      int64                        `json:"planID"`
	ExecutionID int64                        `json:"executionID"`
	RepoNumber  int64                        `json:"repoNumber"`
	PlanNumber  int64                        `json:"planNumber"`
	TriggerId   int64                        `json:"triggerID"`
	TriggerType constant.ScanPlanTriggerType `json:"triggerType"`
	Status      constant.ScanTaskStatus      `json:"status"`
	Result      constant.ScanTaskResult      `json:"result"`
	CreatedBy   string                       `json:"createdBy"`
	CreatedDate *datetime.Local              `json:"createdDate"`
	EditedDate  *datetime.Local              `json:"editedDate"`
	ScanType    constant.ScanType            `json:"scanType"`
}

type ScanTaskReport struct {
	Content string `json:"content"`
}

type ScanTaskInfo struct {
	ID              int64                    `json:"id"`
	RepoID          int64                    `json:"repoID"`
	RepoName        string                   `json:"repoName"`
	PlanID          int64                    `json:"planID"`
	PlanName        string                   `json:"planName"`
	PipelineName    string                   `json:"pipelineName"`
	ExecutionNumber int64                    `json:"executionNumber"`
	ExecutionID     int64                    `json:"executionID"`
	ExecutionRef    string                   `json:"executionRef"`
	RepoNumber      int64                    `json:"repoNumber"`
	PlanNumber      int64                    `json:"planNumber"`
	Trigger         *ScanTaskTrigger         `json:"trigger"`
	Status          constant.ScanTaskStatus  `json:"status"`
	Result          constant.ScanTaskResult  `json:"result"`
	ScanType        constant.ScanType        `json:"scanType"`
	IssueNumber     int64                    `json:"issueNumber"`
	Cost            int64                    `json:"cost"`
	CreatedBy       string                   `json:"createdBy"`
	CreatedDate     *datetime.Local          `json:"createdDate"`
	EditedDate      *datetime.Local          `json:"editedDate"`
	Started         int64                    `json:"started"`
	Finished        int64                    `json:"finished"`
}

func (s *ScanTaskInfo) MetricKind() constant.MetricSourceKind {
	return constant.MetricSourceKindScanTask
}

type ScanTaskTrigger struct {
	ID   int64                        `json:"triggerID,omitempty"`
	Name string                       `json:"triggerName,omitempty"`
	Type constant.ScanPlanTriggerType `json:"triggerType,omitempty"`
}

// ScanTaskInfoReqFilter 查询扫描任务详情过滤条件
type ScanTaskInfoReqFilter struct {
	TaskID     int64                   `json:"taskID,omitempty" form:"taskID"`
	RepoID     int64                   `json:"repoID,omitempty" form:"repoID"`
	PlanID     int64                   `json:"planID,omitempty" form:"planID"`
	TriggerID  int64                   `json:"triggerID,omitempty" form:"triggerID"`
	Branch     string                  `json:"branch,omitempty" form:"branch"`
	Status     constant.ScanTaskStatus `json:"status,omitempty" form:"status"`
	Result     constant.ScanTaskResult `json:"result,omitempty" form:"result"`
	CreateDate string                  `json:"createDate,omitempty" form:"createDate"`
	StartedGt  string                  `json:"startedGt,omitempty" form:"startedGt"`
	StartedLt  string                  `json:"startedLt,omitempty" form:"startedLt"`
	FinishedGt string                  `json:"finishedGt,omitempty" form:"finishedGt"`
	FinishedLt string                  `json:"finishedLt,omitempty" form:"finishedLt"`
	Sort       string                  `json:"sort" form:"sort"`
	Order      string                  `json:"order" form:"order"`
	Page       int                     `json:"page" form:"page" default:"1"`
	PageSize   int                     `json:"pageSize" form:"pageSize" default:"20"`
}

type ScanTaskReqForCreate struct {
	PlanID uint   `json:"planID" binding:"required"`
	RepoID uint   `json:"repoID" binding:"required"`
	Branch string `json:"branch" binding:"required"`
}

// ScanTaskRunnerCreateReq 流水线插件回调创建扫描任务的请求体。
// codescan 插件在非内嵌流水线（INNER_PIPELINE 非 "true"）中扫描前调用
// POST /scan/tasks/create-by-runner 换取任务 ID 与该分支的提交范围。triggerType 键恒存在、
// 值可为空字符串，传 GITFOX_BUILD_TRIGGER 原值（branch_updated/pullreq_created/…），
// 服务端归一为任务表的 manual/cron/action 后落库；不带 triggerID —— 这类任务由普通流水线
// 触发，其触发器与 scan 触发器不同表，triggerID 只由内置 scan 流水线写入。
type ScanTaskRunnerCreateReq struct {
	RepoID      int64  `json:"repoID" binding:"required"`
	PlanID      int64  `json:"planID" binding:"required"`
	ExecutionID int64  `json:"executionID" binding:"required"`
	TriggerType string `json:"triggerType"`
	// Branch 是插件声明的扫描分支（SCAN_BRANCH），服务端据此解析提交范围。
	// 必填：普通流水线必须显式声明扫描分支，缺失或为空直接拒绝。
	Branch string `json:"branch" binding:"required"`
}

// ScanTaskRunnerCreateResp create-by-runner 的响应体。commitBefore/commitAfter 供插件写入
// GITFOX_COMMIT_BEFORE/GITFOX_COMMIT_AFTER 判定扫描范围：两者相同 → 全量；不同 → 增量
// （按 git diff commitBefore 过滤报告行）。取值由扫描计划的 scanType 决定，见
// resolveRunnerCommitRange。
type ScanTaskRunnerCreateResp struct {
	TaskID       int64  `json:"taskID"`
	CommitBefore string `json:"commitBefore"`
	CommitAfter  string `json:"commitAfter"`
}
