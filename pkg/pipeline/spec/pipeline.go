// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

// Pipeline defines the pipeline execution.
type Pipeline struct {
	Stages  []*Stage          `json:"stages,omitempty"`
	Inputs  map[string]*Input `json:"inputs,omitempty"`
	Options *Default          `json:"options,omitempty"`
}
