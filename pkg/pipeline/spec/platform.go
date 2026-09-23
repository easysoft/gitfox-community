// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

// Platform defines the target execution environment.
type Platform struct {
	Arch     string   `json:"arch,omitempty"`
	Features []string `json:"features,omitempty"`
	Os       string   `json:"os,omitempty"`
	Variant  string   `json:"variant,omitempty"`
	Version  string   `json:"version,omitempty"`
}
