package req

import (
	"github.com/gin-gonic/gin"
)

const (
	PathParamRepoRef  = "repo_ref"
	QueryParamRepoID  = "repo_id"
	QueryParamSpaceID = "space_id"
)

func GetRepoRefFromPath(r *gin.Context) (string, error) {
	return PathParamOrError(r, PathParamRepoRef)
}

// TODO: 解析 repoRef 获取 repo 对象的通用方法
