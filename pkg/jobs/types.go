// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package jobs

// StateChange represents a change in the state of a job.
// 状态变更表示作业状态的变更。
type StateChange struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	State    string `json:"state"`
	Progress int    `json:"progress"`
	Result   string `json:"result"`
	Failure  string `json:"failure"`
}
