// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

// TemplateStage defines a stage-level template
type TemplateStage struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Stage       *Stage            `json:"stage,omitempty"`
	Inputs      map[string]*Input `json:"inputs,omitempty"`
}
