package middleware

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Attempt 返回一个 Gin 中间件，用于认证请求
// 如果有认证数据则进行认证，否则使用匿名用户会话
// 使用方式: router.Use(middleware.Attempt(authenticator))
func Attempt(authenticator authn.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetLogger(c)
		ctx := c.Request.Context()

		// 尝试认证请求
		session, err := authenticator.Authenticate(c.Request)
		if err != nil {
			log.Debugf("authentication failed: %v", err)
			resp.Unauthorized(c, err)
			return
		}

		// 更新日志上下文，注入用户信息
		log = log.With(
			"principal_uid", session.Principal.UID,
			"principal_type", string(session.Principal.Type),
			"principal_admin", session.Principal.Admin,
		)
		c.Set(logger.ContextLogger, log)

		// 将认证会话注入到 Context 中
		c.Request = c.Request.WithContext(
			req.SetAuthSession(ctx, session),
		)

		c.Next()
	}
}
