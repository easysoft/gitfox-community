// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

// Mount defines a volume mount.
type Mount struct {
	Name     string `json:"name,omitempty"`
	Path     string `json:"path,omitempty"`
	SubPath  string `json:"subPath,omitempty"`
	ReadOnly bool   `json:"readOnly,omitempty"`
}
