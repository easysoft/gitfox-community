package i18n

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// X-Locale 自定义 Header 名称
const XLocaleHeader = "X-Locale"

// GinMiddleware Gin 中间件
func GinMiddleware(config *MiddlewareConfig) gin.HandlerFunc {

	if config == nil {
		config = GetGlobalManager().ToMiddlewareConfig()
	}

	if config.DefaultLocale == "" {
		config.DefaultLocale = "zh"
	}

	return func(c *gin.Context) {
		locale := detectLocale(c.Request, config)

		// 设置到 context 中
		ctx := WithLocale(c.Request.Context(), locale)
		c.Request = c.Request.WithContext(ctx)

		// 设置到 Gin context 中
		c.Set(XLocaleHeader, locale)

		c.Next()
	}
}

// HTTPMiddleware 标准 HTTP 中间件
func HTTPMiddleware(config *MiddlewareConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = &MiddlewareConfig{}
	}

	if config.DefaultLocale == "" {
		config.DefaultLocale = "en"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			locale := detectLocale(r, config)
			ctx := WithLocale(r.Context(), locale)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	DefaultLocale   string
	AcceptedLocales []string
	QueryParam      string
	CookieName      string
	HeaderName      string
	SessionKey      string
}

// detectLocale 检测语言环境
func detectLocale(r *http.Request, config *MiddlewareConfig) string {
	// 1. 从查询参数获取
	if config.QueryParam != "" {
		if locale := r.URL.Query().Get(config.QueryParam); locale != "" {
			if isAcceptedLocale(locale, config.AcceptedLocales) {
				return locale
			}
		}
	}

	// 2. 从 Cookie 获取
	if config.CookieName != "" {
		if cookie, err := r.Cookie(config.CookieName); err == nil {
			if isAcceptedLocale(cookie.Value, config.AcceptedLocales) {
				return cookie.Value
			}
		}
	}

	// 3. 从自定义 Header 获取
	if config.HeaderName != "" {
		if locale := r.Header.Get(config.HeaderName); locale != "" {
			if isAcceptedLocale(locale, config.AcceptedLocales) {
				return locale
			}
		}
	}

	// 4. 从 Accept-Language Header 获取
	if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
		if locale := parseAcceptLanguage(acceptLang, config.AcceptedLocales); locale != "" {
			return locale
		}
	}

	// 5. 返回默认语言环境
	return config.DefaultLocale
}

// isAcceptedLocale 检查是否为接受的语言环境
func isAcceptedLocale(locale string, acceptedLocales []string) bool {
	if len(acceptedLocales) == 0 {
		return true // 如果没有限制，接受所有语言环境
	}

	for _, accepted := range acceptedLocales {
		if locale == accepted {
			return true
		}
	}
	return false
}

// parseAcceptLanguage 解析 Accept-Language 头
func parseAcceptLanguage(acceptLang string, acceptedLocales []string) string {
	// 简化的 Accept-Language 解析
	languages := strings.Split(acceptLang, ",")

	for _, lang := range languages {
		// 移除权重信息 (q=0.9)
		lang = strings.Split(strings.TrimSpace(lang), ";")[0]

		// 检查完整匹配
		if isAcceptedLocale(lang, acceptedLocales) {
			return lang
		}

		// 检查语言代码匹配 (zh-CN -> zh)
		if parts := strings.Split(lang, "-"); len(parts) > 1 {
			if isAcceptedLocale(parts[0], acceptedLocales) {
				return parts[0]
			}
		}
	}

	return ""
}

// Gin 辅助函数

// GetLocaleFromGin 从 Gin context 获取语言环境
func GetLocaleFromGin(c *gin.Context) string {
	if locale, exists := c.Get(XLocaleHeader); exists {
		if localeStr, ok := locale.(string); ok {
			return localeStr
		}
	}
	return GetLocaleFromContext(c.Request.Context())
}

// TGin Gin 翻译函数
func TGin(c *gin.Context, key I18nKey, args ...interface{}) string {
	locale := GetLocaleFromGin(c)
	return TWithLocale(locale, string(key), args...)
}

// TfGin Gin 格式化翻译函数
func TfGin(c *gin.Context, key string, data interface{}) string {
	locale := GetLocaleFromGin(c)
	translator, err := GetGlobalManager().GetTranslator(locale)
	if err != nil {
		return key
	}
	return translator.Tf(key, data)
}
