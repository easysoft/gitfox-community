// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package trigger

import "github.com/easysoft/gitfox-community/app/constant"

type Trigger struct {
	ID          int64                  `json:"-"`
	PipelineID  int64                  `json:"pipeline_id"`
	Description string                 `json:"description"`
	Type        string                 `json:"trigger_type"`
	Secret      string                 `json:"-"`
	RepoID      int64                  `json:"repo_id"`
	CreatedBy   int64                  `json:"created_by"`
	Disabled    bool                   `json:"disabled"`
	Actions     constant.TriggerAction `json:"actions"`
}
