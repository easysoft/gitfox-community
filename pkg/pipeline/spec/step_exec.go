// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type StepExec struct {
	Shell  string            `json:"shell,omitempty"`
	Script Stringorslice     `json:"script,omitempty"`
	Envs   map[string]string `json:"envs,omitempty"`
}
