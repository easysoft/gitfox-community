// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package settings

type Setting struct {
	SpaceID int    `json:"space_id"`
	RepoID  int    `json:"repo_id"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

type SystemSetting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
