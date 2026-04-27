package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/authorizer"
	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// ArtifactAuthMiddleware returns a middleware that uses Casbin authorizer to enforce artifact access permissions.
func ArtifactAuthMiddleware(authz authorizer.Authorizer, action enum.CasbinAction) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		log := logger.GetLoggerFromGin(c)

		// 1. Get Auth Session
		session, ok := req.GetAuthSession(ctx)
		if !ok || session.User == nil {
			if action == enum.CasbinActionArtifactPull {
				// Allow anonymous pull, or handle inside authorizer
			} else {
				log.Warnf("Unauthorized access attempt to artifact API: %s", c.Request.URL.Path)
				resp.Unauthorized(c, fmt.Errorf("authentication required"))
				c.Abort()
				return
			}
		}

		// 2. Extract Space, Repo and Entity context from Request
		var spaceID int64
		var repoID int64
		if c.ContentType() == "application/json" {
			var req struct {
				SpaceID int64 `json:"spaceId"`
				RepoID  int64 `json:"repoId"`
			}
			// Use ShouldBindBodyWith to keep body available for controllers
			if err := c.ShouldBindBodyWith(&req, binding.JSON); err == nil {
				spaceID = req.SpaceID
				repoID = req.RepoID
			}
		} else if c.ContentType() != "" {
			// For multipart/form-data or urlencoded
			if sID := c.PostForm("spaceId"); sID != "" {
				spaceID, _ = strconv.ParseInt(sID, 10, 64)
			}
			if rID := c.PostForm("repoId"); rID != "" {
				repoID, _ = strconv.ParseInt(rID, 10, 64)
			}
		}

		// 对 GET 请求或 body 未携带 spaceId 时，从 query params 补充
		if spaceID == 0 {
			if sID := c.Query("spaceId"); sID != "" {
				spaceID, _ = strconv.ParseInt(sID, 10, 64)
			}
		}
		if repoID == 0 {
			if rID := c.Query("repoId"); rID != "" {
				repoID, _ = strconv.ParseInt(rID, 10, 64)
			}
		}

		// 构建 entity：优先从 URL 路径参数解析，否则用仓库级 mock entity
		entity := &model.ArtifactEntity{
			Kind: model.ArtifactEntityKindFormat,
			Name: "repository",
		}
		if entityId := c.Param("entityId"); entityId != "" {
			entity = model.ParseEntityKind(entityId)
		}

		// 3. Enforce Authorization
		err := authz.Authorize(ctx, session, spaceID, repoID, entity, action)
		var uid int64
		if session != nil && session.User != nil {
			uid = session.User.ID
		}
		if err != nil {
			log.Warnf("Artifact Authorization failed for user %d on space %d repo %d: %v", uid, spaceID, repoID, err)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "access denied",
			})
			return
		}

		c.Next()
	}
}
