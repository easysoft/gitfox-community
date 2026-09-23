// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package execution

import (
	"github.com/easysoft/gitfox-community/app/constant"
	stageModel "github.com/easysoft/gitfox-community/app/model/stage"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

type ReqForCreate struct {
	PipelineID uint              `json:"pipelineID" binding:"required"`
	Params     map[string]string `json:"params"`
}

type Execution struct {
	ID           uint              `json:"id"`
	PipelineID   uint              `json:"pipelineID"`
	Parent       uint              `json:"parent"`
	Trigger      string            `json:"trigger"`
	Commit       string            `json:"commit"`
	CommitBefore string            `json:"commitBefore"`
	Ref          string            `json:"ref"`
	Params       map[string]string `json:"params"`
	StartedDate  *datetime.Local   `json:"started"`
	FinishedDate *datetime.Local   `json:"finished"`
	Duration     uint              `json:"duration"`
	Status       constant.CIStatus `json:"status"`
	Error        string            `json:"error"`
	Number       uint              `json:"number"`
	Logs         string            `json:"logs"`
	CreatedBy    string            `json:"createdBy"`
	CreatedDate  *datetime.Local   `json:"createdDate"`
	EditedDate   *datetime.Local   `json:"editedDate"`
}

type ExecutionWithStages struct {
	Execution
	Stages []*stageModel.StageWithSteps `json:"stages"`
}

type ExecutionByPr struct {
	ID         uint              `json:"id" gorm:"id"`
	PipelineID uint              `json:"pipelineID" gorm:"pipelineID"`
	Name       string            `json:"name" gorm:"name"`
	Trigger    string            `json:"trigger" gorm:"trigger"`
	Commit     string            `json:"commit" gorm:"commit"`
	Ref        string            `json:"ref" gorm:"ref"`
	Status     constant.CIStatus `json:"status" gorm:"status"`
}

type ExecutionInfoReqFilter struct {
	Commit string `json:"commit"`
	Ref    string `json:"ref"`
}
