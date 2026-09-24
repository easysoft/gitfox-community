// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package step

import (
	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

// Step 流水线步骤执行记录
type Step struct {
	ID         uint              `json:"id"`
	StageID    uint              `json:"stageID"`
	Name       string            `json:"name"`
	Code       string            `json:"code"`
	SN         int               `json:"sn"`
	Image      string            `json:"image"`
	Depends    []string          `json:"depends"`
	GroupID    uint              `json:"groupID"`
	FailPolicy string            `json:"failPolicy"`
	Started    *datetime.Local   `json:"started"`
	Finished   *datetime.Local   `json:"finished"`
	Duration   uint              `json:"duration"`
	Status     constant.CIStatus `json:"status"`
	ErrorCode  int               `json:"errorCode"`
	ErrorLog   string            `json:"errorLog"`
}
