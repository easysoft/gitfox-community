package middleware

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// TokenType 定义 Token 类型
type TokenType string

const (
	TokenTypeSession TokenType = "session"
	TokenTypePAT     TokenType = "pat"
)

// TokenMetadata 包含 Token 相关的元数据信息
type TokenMetadata struct {
	TokenType TokenType
	TokenID   int64
}

// ImpactsAuthorization 实现 Metadata 接口
func (m *TokenMetadata) ImpactsAuthorization() bool {
	return true
}

// BlockSessionToken 阻止使用 session token 进行认证的请求
// 注意：主要用于阻止使用 session token 进行 Git 操作
// 使用方式: router.Use(middleware.BlockSessionToken())
func BlockSessionToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetLogger(c)

		// 只有在认证数据可用且基于 session token 时才进行阻止
		if session, ok := req.GetAuthSession(c.Request.Context()); ok {
			if tokenMetadata, ok := session.Metadata.(*TokenMetadata); ok &&
				tokenMetadata.TokenType == TokenTypeSession {
				log.Warn("blocking git operation - session tokens are not allowed for usage with git")

				err := errors.BadRequest.New("session tokens are not allowed for usage with git")
				err = errors.WithI18nKey(err, i18n.AuthSessionTokenNotAllowed)
				resp.Unauthorized(c, err)
				return
			}
		}

		c.Next()
	}
}
