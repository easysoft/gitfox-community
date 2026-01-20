package req

import (
	"github.com/gin-gonic/gin"
)

const (
	PathParamRepoRef  = "repoRef"
	PathParamFilePath = "filePath"
	QueryParamRepoID  = "repoID"
	QueryParamSpaceID = "spaceID"

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
