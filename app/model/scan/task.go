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
	TriggerId   *int64                       `json:"triggerID"`
	TriggerType constant.ScanPlanTriggerType `json:"triggerType"`
	Status      constant.ScanTaskStatus      `json:"status"`
	Result      *constant.ScanTaskResult     `json:"result"`
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
	Result          *constant.ScanTaskResult `json:"result"`
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
	TaskID    int64                   `json:"taskID,omitempty" form:"taskID"`
	RepoID    int64                   `json:"repoID,omitempty" form:"repoID"`
	PlanID    int64                   `json:"planID,omitempty" form:"planID"`
	TriggerID int64                   `json:"triggerID,omitempty" form:"triggerID"`
	Branch    string                  `json:"branch,omitempty" form:"branch"`
	Status    constant.ScanTaskStatus `json:"status,omitempty" form:"status"`
	Result    constant.ScanTaskResult `json:"result,omitempty" form:"result"`
	Sort      string                  `json:"sort" form:"sort"`
	Order     string                  `json:"order" form:"order"`
	Page      int                     `json:"page" form:"page" default:"1"`
	PageSize  int                     `json:"pageSize" form:"pageSize" default:"20"`
}

type ScanTaskReqForCreate struct {
	PlanID uint   `json:"planID" binding:"required"`
	RepoID uint   `json:"repoID" binding:"required"`
	Branch string `json:"branch" binding:"required"`
}
