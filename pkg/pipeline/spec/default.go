// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type Default struct {
	Clone      *Clone            `json:"clone,omitempty"`
	Repository *Repository       `json:"repository,omitempty"`
	Registry   *Registry         `json:"registry,omitempty"`
	Resources  *Resources        `json:"resources,omitempty"`
	Delegate   Stringorslice     `json:"delegate,omitempty"`
	Barriers   []*Barrier        `json:"barriers,omitempty"`
	Status     *Status           `json:"status,omitempty"`
	Timeout    string            `json:"timeout,omitempty"`
	Envs       map[string]string `json:"envs,omitempty"`
}
