// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package jobs

import "github.com/easysoft/gitfox-community/pkg/utils/datetime"

// State represents state of a background job.
type State string

// State enumeration.
const (
	StateRunning   State = "running"
	StateScheduled State = "scheduled"
	StateFinished  State = "finished"
	StateFailed    State = "failed"
	StateCanceled  State = "canceled"
)

// Priority represents priority of a background job.
type Priority int

// Priority enumeration.
const (
	PriorityNormal   Priority = 0
	PriorityElevated Priority = 1
)

// ScheduleJob 后台任务模型
type ScheduleJob struct {
	ID                  uint            `json:"id"`
	Name                string          `json:"name"`
	Type                string          `json:"type"`
	Priority            Priority        `json:"priority"`
	Data                string          `json:"data"`
	Result              string          `json:"result"`
	MaxDurationSeconds  int             `json:"maxDurationSeconds"`
	MaxRetries          int             `json:"maxRetries"`
	State               State           `json:"state"`
	Scheduled           *datetime.Local `json:"scheduled"`
	TotalExecutions     int             `json:"totalExecutions"`
	RunBy               string          `json:"runBy"`
	RunDeadline         *datetime.Local `json:"runDeadline"`
	RunProgress         int             `json:"runProgress"`
	LastExecuted        *datetime.Local `json:"lastExecuted"`
	IsRecurring         bool            `json:"isRecurring"`
	RecurringCron       string          `json:"recurringCron"`
	ConsecutiveFailures int             `json:"consecutiveFailures"`
	LastFailureError    string          `json:"lastFailureError"`
	GroupID             string          `json:"groupID"`
	CreatedBy           string          `json:"createdBy"`
	CreatedDate         *datetime.Local `json:"createdDate"`
	EditedBy            string          `json:"editedBy"`
	EditedDate          *datetime.Local `json:"editedDate"`
	Deleted             uint8           `json:"deleted"`
}

// CreateJobRequest 创建/更新定时任务的请求参数
type CreateJobRequest struct {
	JobName            string `json:"jobName" binding:"required"`
	JobType            string `json:"jobType" binding:"required"`
	JobData            string `json:"jobData" binding:"required"`
	CronDef            string `json:"cronDef" binding:"required"`
	MaxDurationSeconds int64  `json:"maxDurationSeconds" binding:"required,min=1"`
}

// DeleteJobRequest 删除定时任务的请求参数
type DeleteJobRequest struct {
	JobName string `json:"jobName" binding:"required"`
}
