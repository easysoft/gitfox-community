// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

// Defines a matrix execution strategy.
type Matrix struct {
	Axis        map[string][]string `json:"axis,omitempty"`
	Concurrency int64               `json:"concurrency,omitempty"`
	Include     []map[string]string `json:"include,omitempty"`
	Exclude     []map[string]string `json:"exclude,omitempty"`
}
