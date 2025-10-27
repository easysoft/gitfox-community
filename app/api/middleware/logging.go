package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// 统一日志前缀
const AccessLogPrefix = "api.access"

// RequestID 生成并注入 trace_id 到 Gin Context 和响应头，同时注入带有 trace_id 的 logger 到上下文
func RequestID(log *logger.SugaredMateLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID, _ := gonanoid.Generate("1234567890abcdefghijklmnopqrstuvwxyz", 16)
		c.Set("trace_id", traceID)
		c.Writer.Header().Set("X-Trace-ID", traceID)

		// 将 trace_id 附加到 logger 并注入到上下文中，方便后续层级调用
		requestLogger := log.With("trace_id", traceID)
		c.Set(logger.ContextLogger, requestLogger)

		c.Next()
	}
}

// AccessLog 统一 API 访问日志输出
// 输出字段：http.method, http.url, status_code, response_size_bytes, elapsed_ms, trace_id, client_ip, user_agent
// 可选输出：query, route_params, request_body(截断), response_body(截断)
func AccessLog(logger *logger.SugaredMateLogger) gin.HandlerFunc {
	const logRequestBody = false    // 如需记录请求体，改为 true
	const logResponseBody = false   // 如需记录响应体，改为 true
	const maxLoggedBodyBytes = 2048 // 体积上限，超过将截断

	return func(c *gin.Context) {
		start := time.Now()

		method := c.Request.Method
		url := c.Request.URL.String()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		query := c.Request.URL.RawQuery
		routeParams := c.Params

		// 可选：记录请求体
		var reqBody string
		if logRequestBody {
			// 仅记录 JSON 或表单，其他类型略过
			ct := c.ContentType()
			if strings.Contains(ct, "application/json") || strings.Contains(ct, "application/x-www-form-urlencoded") {
				b, _ := io.ReadAll(io.LimitReader(c.Request.Body, maxLoggedBodyBytes))
				reqBody = string(b)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(b))
			}
		}

		// 可选：包装响应 writer 以截获响应体
		var bw *bytes.Buffer
		if logResponseBody {
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

		if logRequestBody && reqBody != "" {
			fields = append(fields, "request_body", reqBody)
		}
		if logResponseBody && bw != nil && bw.Len() > 0 {
			resp := bw.String()
			if len(resp) > maxLoggedBodyBytes {
				resp = resp[:maxLoggedBodyBytes]
			}
			fields = append(fields, "response_body", resp)
		}

		logger.Infow(AccessLogPrefix, fields...)
	}
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
