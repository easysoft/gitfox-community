package pipeline

import "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"

type Trigger struct {
	ID          int64           `json:"id"`
	PipelineID  int64           `json:"pipelineID"`
	Trigger     []string        `json:"trigger"`
	Cron        string          `json:"cron"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
}

type TriggerFilter struct {
	Page       int `json:"page" form:"page" default:"1"`
	PageSize   int `json:"pageSize" form:"pageSize" default:"20"`
	PipelineID int `json:"pipelineID"`
}
