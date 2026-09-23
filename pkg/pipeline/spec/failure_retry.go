// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type Retry struct {
	Attempts int64           `json:"attempts,omitempty"`
	Interval Durationorslice `json:"interval,omitempty"`
	Failure  *RetryFailure   `json:"failure,omitempty"`
}
