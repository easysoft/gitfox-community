package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/utils/random"
)

// 统一日志前缀
const AccessLogPrefix = "api.access"

// RequestID 生成并注入 trace_id 到 Gin Context 和响应头，同时注入带有 trace_id 的 logger 到上下文
func RequestID(log *logger.SugaredMateLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := random.RandString(16)
		c.Set("trace_id", traceID)
		c.Writer.Header().Set("X-Trace-ID", traceID)

		// 将 trace_id 附加到 logger 并注入到上下文中，方便后续层级调用
		requestLogger := log.With("trace_id", traceID)
		c.Set(logger.ContextLogger, requestLogger)
		ctx := c.Request.Context()
		ctx = logger.WithTraceID(ctx, traceID)
		ctx = context.WithValue(ctx, logger.LoggerContextKey, requestLogger)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// AccessLog 统一 API 访问日志输出
// 输出字段：http.method, http.url, status_code, response_size_bytes, elapsed_ms, trace_id, client_ip, user_agent
// 可选输出：query, route_params, request_body(截断), response_body(截断)
// 是否打印 request_body / response_body 由配置文件 middleware.logRequestBody / logResponseBody 控制。
func AccessLog(logger *logger.SugaredMateLogger, mwConfig config.MiddlewareConfig) gin.HandlerFunc {
	maxLoggedBodyBytes := mwConfig.MaxLoggedBodyBytes
	if maxLoggedBodyBytes <= 0 {
		maxLoggedBodyBytes = 2048
	}

	return func(c *gin.Context) {
		start := time.Now()

		method := c.Request.Method
		url := c.Request.URL.String()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		query := c.Request.URL.RawQuery
		routeParams := c.Params

		// 可选：记录请求体（按用户真实请求的 body 打印）
		var reqBody string
		if mwConfig.LogRequestBody {
			ct := c.ContentType()
			if strings.Contains(ct, "application/json") || strings.Contains(ct, "application/x-www-form-urlencoded") {
				var buf bytes.Buffer
				tr := io.TeeReader(c.Request.Body, &buf)
				b, _ := io.ReadAll(io.LimitReader(tr, int64(maxLoggedBodyBytes)))
				reqBody = string(b)
				c.Request.Body = io.NopCloser(io.MultiReader(&buf, c.Request.Body))
			}
		}

		// 可选：包装响应 writer 以截获响应体
		var bw *bytes.Buffer
		if mwConfig.LogResponseBody {
			bw = &bytes.Buffer{}
			c.Writer = &bodyLogWriter{ResponseWriter: c.Writer, buf: bw}
		}

		c.Next()

		elapsed := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()
		traceID, _ := c.Get("trace_id")
		traceIDStr, _ := traceID.(string)

		// 结构化日志
		fields := []interface{}{
			"http.method", method,
			"http.url", url,
			"status_code", status,
			"response_size_bytes", size,
			"elapsed_ms", elapsed.Milliseconds(),
			"trace_id", traceIDStr,
			"client_ip", clientIP,
			"user_agent", userAgent,
			"query", query,
			"route_params", routeParams,
		}

		if mwConfig.LogRequestBody && reqBody != "" {
			fields = append(fields, "request_body", parseBody(reqBody))
		}
		if mwConfig.LogResponseBody && bw != nil && bw.Len() > 0 {
			resp := bw.String()
			if len(resp) > maxLoggedBodyBytes {
				resp = resp[:maxLoggedBodyBytes]
			}
			fields = append(fields, "response_body", parseBody(resp))
		}

		logger.Infow(AccessLogPrefix, fields...)
	}
}

// parseBody 将请求/响应体解析为 JSON 对象，方便日志聚合工具直接查询字段。
// 解析失败（如 form-urlencoded、截断的 body、非法 JSON）时回退为原始字符串。
func parseBody(s string) interface{} {
	var obj interface{}
	if err := json.Unmarshal([]byte(s), &obj); err == nil {
		return obj
	}
	return s
}

type bodyLogWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	if w.buf != nil {
		w.buf.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	if w.buf != nil {
		w.buf.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}
