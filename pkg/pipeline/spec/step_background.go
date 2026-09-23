// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type StepBackground struct {
	Shell      string            `json:"shell,omitempty"`
	Envs       map[string]string `json:"envs,omitempty"`
	Run        string            `json:"run,omitempty"`
	Script     Stringorslice     `json:"script,omitempty"`
	Container  *Container        `json:"container,omitempty"`
	Mount      []*Mount          `json:"mount,omitempty"`
	Image      string            `json:"image,omitempty"`
	User       string            `json:"user,omitempty"`
	Pull       string            `json:"pull,omitempty"`
	Entrypoint string            `json:"entrypoint,omitempty"`
	Args       []string          `json:"args,omitempty"`
	Ports      []string          `json:"ports,omitempty"`
	Network    string            `json:"network,omitempty"`
	Privileged bool              `json:"privileged,omitempty"`
	Resources  *Resources        `json:"resources,omitempty"`
	Workdir    string            `json:"workdir,omitempty"`
}
