// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type StepPlugin struct {
	Image      string                 `json:"image,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Uses       string                 `json:"uses,omitempty"`
	Connector  string                 `json:"connector,omitempty"`
	Pull       string                 `json:"pull,omitempty"`
	Envs       map[string]string      `json:"envs,omitempty"`
	Reports    []*Report              `json:"reports,omitempty"`
	Privileged bool                   `json:"privileged,omitempty"`
	User       string                 `json:"user,omitempty"`
	Group      string                 `json:"group,omitempty"`
	Network    string                 `json:"network,omitempty"`
	With       map[string]interface{} `json:"with,omitempty"`
	Inputs     map[string]interface{} `json:"inputs,omitempty"`
	Outputs    []string               `json:"outputs,omitempty"`
	Resources  *Resources             `json:"resources,omitempty"`
	Mount      []*Mount               `json:"mount,omitempty"`
}
