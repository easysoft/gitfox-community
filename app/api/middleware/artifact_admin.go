package middleware

import (
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"

	"github.com/gin-gonic/gin"
)

// ArtifactAdminMiddleware 制品库管理接口鉴权中间件。
//
// 此中间件用于保护「权限配置」类接口（SetPermissions / SetEntityPermission 等），
// 使用系统级鉴权取代制品库内部 Casbin 鉴权，规避「先有鸡还是先有蛋」问题：
//
//  1. 请求必须已通过身份认证（session 存在且 User 不为空），否则返回 401。
//  2. 满足以下任一条件即视为管理员：
//     - session.Principal.Admin == true（EasyAuth 中 App=zentao 时设置）
//     - session.User.IsAdmin == true（数据库中标记的系统管理员）
//
// 非管理员用户访问此接口返回 403。
func ArtifactAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		log := logger.GetLoggerFromGin(c)

		session, ok := req.GetAuthSession(ctx)
		if !ok || session == nil || session.User == nil {
			log.Warnf("Unauthenticated access attempt to artifact admin API: %s", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "authentication required",
			})
			return
		}

		// 系统级管理员判断：
		// 1. Principal.Admin：来自 easy authn（App=zentao）
		// 2. User.IsAdmin：数据库中的系统管理员标志
		isAdmin := (session.Principal != nil && session.Principal.Admin) || session.User.IsAdmin
		if !isAdmin {
			log.Warnf("Non-admin user '%s' attempted artifact permission management", session.User.Account)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "artifact permission management requires system admin privilege",
			})
			return
		}

		c.Next()
	}
}
