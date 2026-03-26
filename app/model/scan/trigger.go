package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

// ScanTriggerType 触发器类型
type ScanTriggerType string

const (
	ScanTriggerTypeCron   ScanTriggerType = "cron"
	ScanTriggerTypeManual ScanTriggerType = "manual"
	ScanTriggerTypeEvent  ScanTriggerType = "event"
)

// ScanTriggerReqForCreate 创建触发器请求
type ScanTriggerReqForCreate struct {
	Name        string                   `json:"name" binding:"required,max=100"`
	Desc        string                   `json:"desc"`
	Type        ScanTriggerType          `json:"type"`
	Cron        string                   `json:"cron,omitempty"`
	CronBranch  string                   `json:"cronBranch,omitempty"`
	Keywords    string                   `json:"keywords,omitempty"`
	Actions     []constant.TriggerAction `json:"actions,omitempty"`
	ScanType    constant.ScanType        `json:"scanType,omitempty"`
	Disabled    bool                     `json:"disabled"`
	SolutionIDs []int64                  `json:"solutionIDs,omitempty"`
}

// ScanTriggerReqForUpdate 更新触发器请求
type ScanTriggerReqForUpdate struct {
	Name        string                   `json:"name,omitempty"`
	Desc        string                   `json:"desc,omitempty"`
	Type        ScanTriggerType          `json:"type,omitempty"`
	Cron        string                   `json:"cron,omitempty"`
	CronBranch  string                   `json:"cronBranch,omitempty"`
	Keywords    string                   `json:"keywords,omitempty"`
	Actions     []constant.TriggerAction `json:"actions,omitempty"`
	ScanType    constant.ScanType        `json:"scanType,omitempty"`
	Disabled    *bool                    `json:"disabled,omitempty"`
	SolutionIDs []int64                  `json:"solutionIDs,omitempty"`
}

// ScanTriggerReqFilter 触发器列表查询过滤条件
type ScanTriggerReqFilter struct {
	Sort     string `json:"sort" form:"sort" default:"id"`
	Order    string `json:"order" form:"order" default:"desc"`
	Page     int    `json:"page" form:"page" default:"1"`
	PageSize int    `json:"pageSize" form:"pageSize" default:"30"`
}

// ScanTrigger 触发器响应模型
type ScanTrigger struct {
	ID             int64                    `json:"id"`
	SpaceID        int64                    `json:"spaceID"`
	PlanID         int64                    `json:"planID"`
	RepoID         int64                    `json:"repoID"`
	Name           string                   `json:"name"`
	Desc           string                   `json:"desc"`
	Type           ScanTriggerType          `json:"type"`
	ScanType       constant.ScanType        `json:"scanType"`
	Cron           string                   `json:"cron"`
	CronBranch     string                   `json:"cronBranch"`
	Keywords       string                   `json:"keywords"`
	Actions        []constant.TriggerAction `json:"actions"`
	Disabled       bool                     `json:"disabled"`
	CreatedBy      string                   `json:"createdBy"`
	CreatedDate    *datetime.Local          `json:"createdDate"`
	EditedBy       string                   `json:"editedBy"`
	EditedDate     *datetime.Local          `json:"editedDate"`
	Deleted        bool                     `json:"deleted"`
	ConditionIDs   []int64                  `json:"conditionIDs"`
	Conditions     []*ScanCondition         `json:"conditions"`
	ConditionCount int                      `json:"conditionCount"`
	SolutionIDs    []int64                  `json:"solutionIDs"`
	Solutions      []*Solution              `json:"solutions"`
	SolutionCount  int                      `json:"solutionCount"`

	LatestTaskID      *int64                   `json:"latestTaskID"`
	LatestTaskCreated *datetime.Local          `json:"latestTaskCreated"`
	LatestTaskResult  *constant.ScanTaskResult `json:"latestTaskResult"`
	LatestTaskStatus  *constant.ScanTaskStatus `json:"latestTaskStatus"`
}
