package pipeline

import (
	"fmt"
	"path"
)

type S3Logs struct {
	PipelineID  string `json:"pipelineID"`
	ExecutionID string `json:"executionID"`
	StageID     string `json:"stageID"`
	StepID      string `json:"stepID"`
}

// key 生成 S3 对象路径: {prefix}/YYMMDD/pipelineID/executionID/stageID/stepID.log
func (s *S3Logs) Key() string {
	return path.Join("/", s.PipelineID, s.ExecutionID, s.StageID, fmt.Sprintf("%s.log", s.StepID))
}
