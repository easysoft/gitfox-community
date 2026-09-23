// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

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

// key 生成 S3 对象路径: /pipelineID/executionID/stageID/stepID.log
func (s *S3Logs) Key() string {
	return path.Join("/", s.PipelineID, s.ExecutionID, s.StageID, fmt.Sprintf("%s.log", s.StepID))
}
