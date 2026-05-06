package middleware

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"strconv"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/authorizer"
	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// ArtifactAuthMiddleware returns a middleware that uses Casbin authorizer to enforce artifact access permissions.
// 该中间件使用 Casbin 授权器强制执行工访问权限。
// 1. 获取认证会话
// 2. 从请求中提取空间、存储库和实体上下文
// 3. 当 artifactID > 0 时，优先解析实际 view 所属的 space/repo 作用域
// 3. 系统管理员短路放行
// 4. 执行授权
func ArtifactAuthMiddleware(authz authorizer.Authorizer, artSvc artifactSvc.Interface, action enum.CasbinAction) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		log := logger.GetLoggerFromGin(c)

		// 1. 获取认证会话
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

		// 2. 从请求中提取空间、存储库和实体上下文
		var spaceID int64
		var repoID int64
		var artifactID int64
		if c.ContentType() == "application/json" {
			var req struct {
				SpaceID    int64 `json:"spaceId"`
				RepoID     int64 `json:"repoId"`
				ArtifactID int64 `json:"artifactID"`
			}
			// Use ShouldBindBodyWith to keep body available for controllers
			if err := c.ShouldBindBodyWith(&req, binding.JSON); err == nil {
				spaceID = req.SpaceID
				repoID = req.RepoID
				artifactID = req.ArtifactID
			}
		} else if c.ContentType() != "" {
			// For multipart/form-data or urlencoded
			if sID := c.PostForm("spaceId"); sID != "" {
				spaceID, _ = strconv.ParseInt(sID, 10, 64)
			}
			if rID := c.PostForm("repoId"); rID != "" {
				repoID, _ = strconv.ParseInt(rID, 10, 64)
			}
			if vID := c.PostForm("artifactID"); vID != "" {
				artifactID, _ = strconv.ParseInt(vID, 10, 64)
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
		if artifactID == 0 {
			if vID := c.Query("artifactID"); vID != "" {
				artifactID, _ = strconv.ParseInt(vID, 10, 64)
			}
		}

		if artifactID > 0 && artSvc != nil {
			view, err := artSvc.ResolveView(ctx, &model.SpaceRepoReq{
				SpaceID:    spaceID,
				RepoID:     repoID,
				ArtifactID: artifactID,
			})
			if err != nil {
				resp.BadRequest(c, err)
				c.Abort()
				return
			}
			spaceID = view.SpaceID
			repoID = view.RepoID
		}

		// 构建 entity：优先从 URL 路径参数解析，否则用仓库级 mock entity
		entity := &model.ArtifactEntity{
			Kind: model.ArtifactEntityKindFormat,
			Name: "repository",
		}
		if entityId := c.Param("entityId"); entityId != "" {
			entity = model.ParseEntityKind(entityId)
		}

		// 4. 执行授权
		err := authz.Authorize(ctx, session, spaceID, repoID, entity, action)
		var uid int64
		if session != nil && session.User != nil {
			uid = session.User.ID
		}
		if err != nil {
			log.Warnf("Artifact Authorization failed for user %d on space %d repo %d: %v", uid, spaceID, repoID, err)
			switch {
			case err.Error() == "authentication required":
				err = pkgErrors.WithI18nKey(err, i18n.ArtifactAuthRequired)
			case stderrors.Is(err, rbac.ErrNotAuthorized), err.Error() == "rbac authorize failed":
				err = pkgErrors.WithI18nKey(err, i18n.ArtifactPermissionDenied)
			}
			err = pkgErrors.WithHttpStatus(err, http.StatusForbidden)
			resp.Error(c, err)
			c.Abort()
			return
		}

		c.Next()
	}
}
