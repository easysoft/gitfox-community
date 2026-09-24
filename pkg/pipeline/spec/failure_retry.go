// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type Retry struct {
	Attempts int64           `json:"attempts,omitempty"`
	Interval Durationorslice `json:"interval,omitempty"`
	Failure  *RetryFailure   `json:"failure,omitempty"`
}
