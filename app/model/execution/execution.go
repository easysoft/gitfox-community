package execution

import "time"

type ReqForCreate struct {
	RepoID     uint              `json:"repoID" binding:"required"`
	PipelineID uint              `json:"pipelineID" binding:"required"`
	Branch     string            `json:"branch"`
	Params     map[string]string `json:"params"`
}

type Execution struct {
	ID          uint       `json:"id"`
	PipelineID  uint       `json:"pipelineID"`
	Parent      uint       `json:"parent"`
	Trigger     string     `json:"trigger"`
	Commit      string     `json:"commit"`
	Ref         string     `json:"ref"`
	Params      string     `json:"params"`
	Started     *time.Time `json:"started"`
	Finished    *time.Time `json:"finished"`
	Duration    uint       `json:"duration"`
	Status      string     `json:"status"`
	Error       string     `json:"error"`
	CreatedBy   string     `json:"createdBy"`
	CreatedDate *time.Time `json:"createdDate"`
	EditedDate  *time.Time `json:"editedDate"`
}
