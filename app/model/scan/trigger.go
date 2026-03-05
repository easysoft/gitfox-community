package scan

import (
	"time"
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
	Name        string          `json:"name" binding:"required,max=100"`
	Desc        string          `json:"desc"`
	Type        ScanTriggerType `json:"type"`
	Cron        string          `json:"cron,omitempty"`
	CronBranch  string          `json:"cronBranch,omitempty"`
	Keywords    string          `json:"keywords,omitempty"`
	Actions     string          `json:"actions,omitempty"`
	Disabled    bool            `json:"disabled"`
}

// ScanTriggerReqForUpdate 更新触发器请求
type ScanTriggerReqForUpdate struct {
	Name       string          `json:"name,omitempty"`
	Desc       string          `json:"desc,omitempty"`
	Type       ScanTriggerType `json:"type,omitempty"`
	Cron       string          `json:"cron,omitempty"`
	CronBranch string          `json:"cronBranch,omitempty"`
	Keywords   string          `json:"keywords,omitempty"`
	Actions    string          `json:"actions,omitempty"`
	Disabled   *bool           `json:"disabled,omitempty"`
}

// ScanTrigger 触发器响应模型
type ScanTrigger struct {
	ID          int64           `json:"id"`
	PlanID      int64           `json:"planID"`
	RepoID      int64           `json:"repoID"`
	Name        string          `json:"name"`
	Desc        string          `json:"desc"`
	Type        ScanTriggerType `json:"type"`
	Cron        string          `json:"cron,omitempty"`
	CronBranch  string          `json:"cronBranch,omitempty"`
	Keywords    string          `json:"keywords,omitempty"`
	Actions     string          `json:"actions,omitempty"`
	Disabled    bool            `json:"disabled"`
	CreatedBy   string          `json:"createdBy,omitempty"`
	CreatedDate *time.Time      `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *time.Time      `json:"editedDate"`
	Deleted     bool            `json:"deleted"`
}
