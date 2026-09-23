// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

// Cache defines pipeline caching behavior.
type Cache struct {
	Enabled bool     `json:"enabled,omitempty"`
	Paths   []string `json:"paths,omitempty"`
	Key     string   `json:"key,omitempty"`
	Policy  string   `json:"policy,omitempty"`
}
