package middleware

import (
	"net/http"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

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
			// 对于 Git 请求，如果没有认证数据，返回 401 + WWW-Authenticate header
			// 这会触发 Git 客户端提示用户输入账号密码
			path := c.Request.URL.Path
			if errors.Is(err, constant.ErrAuthNoData) && isGitRequest(path) {
				renderBasicAuth(c)
				return
			}

			// 其他认证失败情况
			log.Debugf("authentication failed: %v", err)
			err = errors.WithI18nKey(err, i18n.AuthFailed)
			resp.Unauthorized(c, err)
			return
		}

		// 更新日志上下文，注入用户信息
		if session.User != nil {
			log = log.With("principal_account", session.User.Account)
			c.Set(logger.ContextLogger, log)
		}

		// 将认证会话注入到 Context 中
		c.Request = c.Request.WithContext(
			req.SetAuthSession(ctx, session),
		)

		c.Next()
	}
}

// isGitRequest 检查是否是 Git 协议相关的请求
func isGitRequest(path string) bool {
	return len(path) > 0 && (strings.Contains(path, "/info/refs") || strings.Contains(path, "/git-upload-pack") || strings.Contains(path, "/git-receive-pack"))
}

// renderBasicAuth renders a response that indicates that the client (GIT) requires basic authentication.
// This is required in order to tell git CLI to query user credentials.
func renderBasicAuth(c *gin.Context) {
	// 设置 WWW-Authenticate header，告诉 Git 客户端需要 Basic 认证
	// realm 可以是任意字符串，Git 客户端会显示给用户
	c.Header("WWW-Authenticate", `Basic realm="GitFox Repository"`)
	c.AbortWithStatus(http.StatusUnauthorized)
}

func MockAttempt() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetLogger(c)
		ctx := c.Request.Context()

		// 尝试认证请求

		session := &auth.Session{
			Principal: &principal.Principal{
				App: "devops",
			},
			User: &principal.User{
				ID:      2,
				Email:   "gitfox@chandao.com",
				Account: "gitfox",
			},
		}

		// 更新日志上下文，注入用户信息
		log = log.With("principal_account", session.User.Account)
		c.Set(logger.ContextLogger, log)

		// 将认证会话注入到 Context 中
		c.Request = c.Request.WithContext(
			req.SetAuthSession(ctx, session),
		)

		c.Next()
	}
}
