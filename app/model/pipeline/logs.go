package pipeline

type S3Logs struct {
	PipelineID  string `json:"pipelineID"`
	ExecutionID string `json:"executionID"`
	StageID     string `json:"stageID"`
	StepID      string `json:"stepID"`
}
