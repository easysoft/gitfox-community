// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package req

import (
	"github.com/gin-gonic/gin"
)

const (
	PathParamRepoRef      = "repoRef"
	PathParamFilePath     = "filePath"
	QueryParamRepoName    = "repoName"
	QueryParamBranchName  = "branchName"
	QueryParamRepoID      = "repoID"
	QueryParamSpaceCode   = "spaceCode"
	QueryParamSpaceID     = "spaceID"

	QueryParamOperation = "Operation"

	TargetPathToSourcePath = "targetPathToSourcePath"
	PullrequestID          = "pullrequestid"
	QueryParamPage         = "page"
	QueryParamLimit        = "limit"
)

func GetRepoRefFromPath(r *gin.Context) (string, error) {
	return PathParamOrError(r, PathParamRepoRef)
}

// TODO: 解析 repoRef 获取 repo 对象的通用方法
