package pipeline

type S3Logs struct {
	Date        string `json:"date"` //// 日期，格式为 YYYYMMDD eg 20230801
	PipelineID  string `json:"pipelineID"`
	ExecutionID string `json:"executionID"`
	StageID     string `json:"stageID"`
	StepID      string `json:"stepID"`
}
