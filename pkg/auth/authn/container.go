package authn

import (
	"fmt"
	"net/http"

	requtil "github.com/easysoft/gitfox-community/app/api/utils/req"
	ctnAdapter "github.com/easysoft/gitfox-community/app/artifact/adapter/container"
	"github.com/easysoft/gitfox-community/app/constant"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// ContainerRegistryBasicAuth 为 OCI Registry 路由提供 Basic 认证。
// 成功时注入 auth session，失败时按 OCI 规范返回 401 错误体。
func ContainerRegistryBasicAuth(realm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetLoggerFromGin(c)

		authenticator := Get(AuthKindBasic)
		if authenticator == nil {
			log.Errorf("【container auth】basic authenticator not registered")
			renderContainerBasicAuthErr(c, realm, "basic authenticator is not registered")
			return
		}

		session, err := authenticator.Authenticate(c)
		if err != nil {
			if pkgErrors.Is(err, constant.ErrAuthNoData) {
				renderContainerBasicAuthErr(c, realm, "authentication required")
				return
			}
			log.Warnf("【container auth】invalid credential on %s %s", c.Request.Method, c.Request.URL.Path)
			renderContainerBasicAuthErr(c, realm, "invalid username or password")
			return
		}

		ctx := requtil.SetAuthSession(c.Request.Context(), session)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func renderContainerBasicAuthErr(c *gin.Context, realm, message string) {
	c.Header("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"errors": []*ctnAdapter.RegistryErr{
			ctnAdapter.ErrUnauthorized.WithDetail(message),
		},
	})
}
