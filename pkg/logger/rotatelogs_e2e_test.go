// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/easysoft/gitfox-community/pkg/config"
)

func TestE2ERotateByDate(t *testing.T) {
	logDir := t.TempDir()
	cfg := config.LoggingConfig{
		LogDir:         logDir,
		Level:          0,
		Stdout:         false,
		Encoder:        "json",
		MaxAge:         30,
		MaxCount:       20,
		MaxSize:        500,
		TimeFormat:     "2006-01-02 15:04:05.999",
		LogDefaultConf: config.LogDefaultConfig{Filename: "default.log"},
		LogSQLConf:     config.LogSQLConfig{Filename: "sql.log"},
	}

	// zap 日志
	l := NewZapMateLogger(cfg, LogTypeDefault).Sugar()
	l.Info("hello zap date rotation")
	l.Flush()

	// SQL 日志
	w := NewSQLFileWriter(cfg)
	_, _ = w.Write([]byte("SQL statement test\n"))

	today := time.Now().Format("2006-01-02")

	// 1. 日期文件名存在
	dateFile := filepath.Join(logDir, "default.log."+today)
	if _, err := os.Stat(dateFile); err != nil {
		t.Fatalf("expected dated file %s, got err %v", dateFile, err)
	}
	sqlDateFile := filepath.Join(logDir, "sql.log."+today)
	if _, err := os.Stat(sqlDateFile); err != nil {
		t.Fatalf("expected dated sql file %s, err %v", sqlDateFile, err)
	}

	// 2. 内容写入日期文件
	data, err := os.ReadFile(dateFile)
	if err != nil {
		t.Fatalf("failed to read dated file: %v", err)
	}
	if !strings.Contains(string(data), "hello zap date rotation") {
		t.Fatalf("dated file missing content: %s", string(data))
	}

	// 3. sql 内容
	sqlData, err := os.ReadFile(sqlDateFile)
	if err != nil {
		t.Fatalf("failed to read dated sql file: %v", err)
	}
	if !strings.Contains(string(sqlData), "SQL statement test") {
		t.Fatalf("dated sql file missing content: %s", string(sqlData))
	}
}
