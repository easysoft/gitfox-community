// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package logger

import (
	"sync"

	"github.com/easysoft/gitfox-community/pkg/config"
)

var (
	globalLog     *SugaredMateLogger
	globalLogOnce sync.Once
)

// InitGlobalLogger 初始化全局 logger（写入 default.log）。
// 供无 context 的模块（git 操作、webhook、pipeline 等）调用，
// 需在应用启动时尽早调用（NewSystemLogger 内）。
// 注意：system.log 仅用于系统启动/停止日志（SystemLog），业务日志统一走 default.log。
func InitGlobalLogger(cfg config.LoggingConfig) {
	globalLogOnce.Do(func() {
		globalLog = NewZapMateLogger(cfg, LogTypeDefault).SetAsyncer(1000).Sugar()
	})
}

// getGlobalLog 返回全局 logger；未初始化时回退到默认 logger，避免 panic
func getGlobalLog() *SugaredMateLogger {
	if globalLog != nil {
		return globalLog
	}
	return getDefaultLogger()
}

// ---- 包级语法糖函数，统一模块日志调用入口 ----

// Trace 记录一条 trace 级别的日志（映射到 debug）。
func Trace(args ...interface{}) {
	getGlobalLog().Trace(args...)
}

// Tracef 使用格式字符串记录一条 trace 级别的日志。
func Tracef(template string, args ...interface{}) {
	getGlobalLog().Tracef(template, args...)
}

// Debug 记录一条 debug 级别的日志。
func Debug(args ...interface{}) {
	getGlobalLog().Debug(args...)
}

// Debugf 使用格式字符串记录一条 debug 级别的日志。
func Debugf(template string, args ...interface{}) {
	getGlobalLog().Debugf(template, args...)
}

// Debugw 记录一条带结构化字段的 debug 级别日志。
func Debugw(msg string, keysAndValues ...interface{}) {
	getGlobalLog().Debugw(msg, keysAndValues...)
}

// Info 记录一条 info 级别的日志。
func Info(args ...interface{}) {
	getGlobalLog().Info(args...)
}

// Infof 使用格式字符串记录一条 info 级别的日志。
func Infof(template string, args ...interface{}) {
	getGlobalLog().Infof(template, args...)
}

// Infow 记录一条带结构化字段的 info 级别日志。
func Infow(msg string, keysAndValues ...interface{}) {
	getGlobalLog().Infow(msg, keysAndValues...)
}

// Warn 记录一条 warn 级别的日志。
func Warn(args ...interface{}) {
	getGlobalLog().Warn(args...)
}

// Warnf 使用格式字符串记录一条 warn 级别的日志。
func Warnf(template string, args ...interface{}) {
	getGlobalLog().Warnf(template, args...)
}

// Warnw 记录一条带结构化字段的 warn 级别日志。
func Warnw(msg string, keysAndValues ...interface{}) {
	getGlobalLog().Warnw(msg, keysAndValues...)
}

// Error 记录一条 error 级别的日志。
func Error(args ...interface{}) {
	getGlobalLog().Error(args...)
}

// Errorf 使用格式字符串记录一条 error 级别的日志。
func Errorf(template string, args ...interface{}) {
	getGlobalLog().Errorf(template, args...)
}

// Errorw 记录一条带结构化字段的 error 级别日志。
func Errorw(msg string, keysAndValues ...interface{}) {
	getGlobalLog().Errorw(msg, keysAndValues...)
}
