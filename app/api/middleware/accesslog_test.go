// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

func TestAccessLogBodyConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name     string
		mw       config.MiddlewareConfig
		wantReq  bool
		wantResp bool
	}{
		{"both on", config.MiddlewareConfig{LogRequestBody: true, LogResponseBody: true, MaxLoggedBodyBytes: 2048}, true, true},
		{"both off", config.MiddlewareConfig{LogRequestBody: false, LogResponseBody: false}, false, false},
		{"req only", config.MiddlewareConfig{LogRequestBody: true, LogResponseBody: false}, true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logDir := t.TempDir()
			cfg := config.LoggingConfig{
				LogDir:         logDir,
				Level:          0,
				Stdout:         false,
				Encoder:        "json",
				MaxSize:        5,
				MaxCount:       3,
				MaxAge:         1,
				TimeFormat:     "2006-01-02 15:04:05.999",
				LogDefaultConf: config.LogDefaultConfig{Filename: "default.log"},
			}
			l := logger.NewZapMateLogger(cfg, logger.LogTypeDefault).Sugar()

			r := gin.New()
			r.Use(AccessLog(l, tc.mw))
			r.POST("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"ok": true})
			})

			reqBody := `{"name":"alice","age":30}`
			req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			today := time.Now().Format("2006-01-02")
			data, err := os.ReadFile(filepath.Join(logDir, "default.log."+today))
			if err != nil {
				t.Fatalf("failed to read log: %v", err)
			}
			out := string(data)

			if tc.wantReq && (!strings.Contains(out, `"request_body":{`) || !strings.Contains(out, `"name":"alice"`) || !strings.Contains(out, `"age":30`)) {
				t.Fatalf("expected request_body as json object, got: %s", out)
			}
			if tc.wantReq && strings.Contains(out, `"request_body":"{`) {
				t.Fatalf("expected request_body not escaped string, got: %s", out)
			}
			if !tc.wantReq && strings.Contains(out, "request_body") {
				t.Fatalf("expected no request_body, got: %s", out)
			}
			if tc.wantResp && !strings.Contains(out, `"response_body":{"ok":true}`) {
				t.Fatalf("expected response_body as json object, got: %s", out)
			}
			if tc.wantResp && strings.Contains(out, `"response_body":"{`) {
				t.Fatalf("expected response_body not escaped string, got: %s", out)
			}
			if !tc.wantResp && strings.Contains(out, "response_body") {
				t.Fatalf("expected no response_body, got: %s", out)
			}
		})
	}
}
