// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package database

import (
	"io"
	"os"

	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

// buildGormWriter 构建 GORM 日志输出 writer。
// writeToFile 为 true 时始终写入 sql.log 文件（保证慢 SQL/错误在 printSQL=false 时也能落地），
// stdout 开启时同时输出到终端。
func buildGormWriter(logConf config.LoggingConfig, writeToFile bool) io.Writer {
	var writers []io.Writer
	if writeToFile {
		writers = append(writers, logger.NewSQLFileWriter(logConf))
	}
	if logConf.Stdout {
		writers = append(writers, os.Stdout)
	}
	switch len(writers) {
	case 0:
		return io.Discard
	case 1:
		return writers[0]
	default:
		return io.MultiWriter(writers...)
	}
}
