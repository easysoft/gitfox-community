package req

import (
	"github.com/gin-gonic/gin"
)

const (
	PathParamRemainder = "*"

	PathParamRepoRef  = "repoRef"
	PathParamOther    = "otherPath"
	PathParamFilePath = "filePath"
	QueryParamRepoID  = "repoId"
	QueryParamSpaceID = "spaceId"
)

func GetRepoRefFromPath(r *gin.Context) (string, error) {
	return PathParamOrError(r, PathParamRepoRef)
}

// TODO: 解析 repoRef 获取 repo 对象的通用方法
