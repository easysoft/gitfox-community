package execution

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

type ReqForCreate struct {
	PipelineID uint              `json:"pipelineID" binding:"required"`
	Params     map[string]string `json:"params"`
}

type Execution struct {
	ID          uint            `json:"id"`
	PipelineID  uint            `json:"pipelineID"`
	Parent      uint            `json:"parent"`
	Trigger     string          `json:"trigger"`
	Commit      string          `json:"commit"`
	Ref         string          `json:"ref"`
	Params      string          `json:"params"`
	Started     *datetime.Local `json:"started"`
	Finished    *datetime.Local `json:"finished"`
	Duration    uint            `json:"duration"`
	Status      string          `json:"status"`
	Error       string          `json:"error"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedDate  *datetime.Local `json:"editedDate"`
}
