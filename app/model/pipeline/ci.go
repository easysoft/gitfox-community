// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package pipeline

type CiCacheMetadata struct {
	Key         string `json:"key"`
	Size        int64  `json:"size"`
	ExecutionID int64  `json:"execution_id"`
	Repo        string `json:"repo"`
	RepoId      int64  `json:"repo_id"`
	Space       string `json:"space"`
	SpaceId     int64  `json:"space_id"`
	CreatedAt   int64  `json:"create_at"`
	PulledAt    int64  `json:"pulled_at"`
	Downloads   int    `json:"downloads"`
}
