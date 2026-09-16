package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/easysoft/gitfox-community/pkg/utils/slice"
	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowOriginFunc  func(origin string) bool
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Requested-With"},
		ExposeHeaders:    []string{"X-Trace-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

func CORS(cfg *CORSConfig) gin.HandlerFunc {
	if cfg == nil {
		cfg = DefaultCORSConfig()
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// 非跨域请求直接继续
		if origin == "" {
			c.Next()
			return
		}

		allowOrigin := ""

		if cfg.AllowOriginFunc != nil {
			if cfg.AllowOriginFunc(origin) {
				allowOrigin = origin
			}
		} else if len(cfg.AllowedOrigins) == 0 || slice.Contain(cfg.AllowedOrigins, "*") {
			allowOrigin = "*"
		} else if slice.Contain(cfg.AllowedOrigins, origin) {
			allowOrigin = origin
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)

			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			if len(cfg.ExposeHeaders) > 0 {
				c.Header("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ","))
			}
		}

		// 预检请求处理
		if c.Request.Method == http.MethodOptions {
			if len(cfg.AllowedMethods) > 0 {
				c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ","))
			}
			if len(cfg.AllowedHeaders) > 0 {
				c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ","))
			}
			if cfg.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", strconv.FormatInt(int64(cfg.MaxAge/time.Second), 10))
			}

			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
