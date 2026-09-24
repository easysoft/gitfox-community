// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type StepBuildCommon struct {
	Script    Stringorslice     `json:"script,omitempty"`
	Container *Container        `json:"container,omitempty"`
	Envs      map[string]string `json:"envs,omitempty"`
	Mount     []*Mount          `json:"mount,omitempty"`
	Caching   *Caching          `json:"caching,omitempty"`
	Mirror    Stringorslice     `json:"mirror,omitempty"`
	Proxy     *Proxy            `json:"proxy,omitempty"`
	Version   string            `json:"version,omitempty"`
	BuildEnv  string            `json:"buildEnv,omitempty"`
}
