package step

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

// Step 流水线步骤执行记录
type Step struct {
	ID         uint              `json:"id"`
	StageID    uint              `json:"stageID"`
	Name       string            `json:"name"`
	Code       string            `json:"code"`
	SN         int               `json:"sn"`
	Image      string            `json:"image"`
	Depends    string            `json:"depends"`
	GroupID    uint              `json:"groupID"`
	FailPolicy string            `json:"failPolicy"`
	Started    *datetime.Local   `json:"started"`
	Finished   *datetime.Local   `json:"finished"`
	Duration   uint              `json:"duration"`
	Status     constant.CIStatus `json:"status"`
	ErrorCode  int               `json:"errorCode"`
	ErrorLog   string            `json:"errorLog"`
}
