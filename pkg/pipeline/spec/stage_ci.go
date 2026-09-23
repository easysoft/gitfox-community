// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type StageCI struct {
	Cache    *Cache            `json:"cache,omitempty"`
	Clone    *CloneStage       `json:"clone,omitempty"`
	Platform *Platform         `json:"platform,omitempty"`
	Runtime  *Runtime          `json:"runtime,omitempty"`
	Steps    []*Step           `json:"steps,omitempty"`
	Envs     map[string]string `json:"envs,omitempty"`
	Volumes  []*Volume         `json:"volumes,omitempty"`
	Vars     []*Variable       `json:"vars,omitempty"`
}
