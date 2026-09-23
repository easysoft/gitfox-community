// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package stage

import (
	"github.com/easysoft/gitfox-community/app/constant"
	stepModel "github.com/easysoft/gitfox-community/app/model/step"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

// Stage 流水线执行阶段
type Stage struct {
	ID           uint              `json:"id"`
	RepoID       uint              `json:"repoID"`
	ExecutionID  uint              `json:"executionID"`
	Parent       uint              `json:"parent"`
	Name         string            `json:"name"`
	Code         string            `json:"code"`
	SN           int               `json:"sn"`
	Type         string            `json:"type"`
	Host         string            `json:"host"`
	OS           string            `json:"os"`
	Arch         string            `json:"arch"`
	Labels       map[string]string `json:"labels"`
	Depends      []string          `json:"depends"`
	RunPolicy    string            `json:"runPolicy"`
	CustomPolicy string            `json:"customPolicy"`
	FailPolicy   string            `json:"failPolicy"`
	Started      *datetime.Local   `json:"started"`
	Finished     *datetime.Local   `json:"finished"`
	Duration     uint              `json:"duration"`
	Status       constant.CIStatus `json:"status"`
	ErrorType    string            `json:"errorType"`
	ErrorCode    int               `json:"errorCode"`
	ErrorLog     string            `json:"errorLog"`
}

// StageWithSteps
type StageWithSteps struct {
	Stage
	Steps []*stepModel.Step `json:"steps"`
}
