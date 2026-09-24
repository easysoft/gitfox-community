// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package importer

import (
	"strconv"
	"strings"
)

const jobIDPrefix = "import-repo-"

func JobIDFromRepoID(repoID int64) string {
	return jobIDPrefix + strconv.FormatInt(repoID, 10)
}

func RepoIDFromJobID(jobID string) int64 {
	if !strings.HasPrefix(jobID, jobIDPrefix) {
		return 0
	}
	repoID, _ := strconv.ParseInt(jobID[len(jobIDPrefix):], 10, 64)
	return repoID
}
