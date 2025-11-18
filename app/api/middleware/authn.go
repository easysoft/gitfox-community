package middleware

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Attempt 返回一个 Gin 中间件，用于认证请求
// 使用方式: router.Use(middleware.Attempt(authenticator))
func Attempt(authenticator authn.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetLogger(c)
		ctx := c.Request.Context()

		// 尝试认证请求
		session, err := authenticator.Authenticate(c)
		if err != nil {
			log.Debugf("authentication failed: %v", err)
			resp.Unauthorized(c, err)
			return
		}

		// 更新日志上下文，注入用户信息
		log = log.With("principal_uid", session.User.UID)
		c.Set(logger.ContextLogger, log)

		// 将认证会话注入到 Context 中
		c.Request = c.Request.WithContext(
			req.SetAuthSession(ctx, session),
		)

		c.Next()
	}
}

func MockAttempt() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetLogger(c)
		ctx := c.Request.Context()

		// 尝试认证请求

		session := &auth.Session{
			Principal: principal.User{
				ID:    1,
				Email: "actor@chandao.com",
				UID:   "actor",
			},
			User: &principal.User{
				ID:    2,
				Email: "gitfox@chandao.com",
				UID:   "gitfox",
			},
		}

		// 更新日志上下文，注入用户信息
		log = log.With("principal_uid", session.User.UID)
		c.Set(logger.ContextLogger, log)

		// 将认证会话注入到 Context 中
		c.Request = c.Request.WithContext(
			req.SetAuthSession(ctx, session),
		)

		c.Next()
	}
}
