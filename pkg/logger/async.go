// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logEntry struct {
	fields []zap.Field
	entry  *zapcore.CheckedEntry
}

var logMsgPool *sync.Pool

// Note: func setAsync is must be setting on the root node, Otherwise it will cause other errors!
// 注意: func setAsync必须在根节点上设置，否则会导致其他错误!
// Child node cannot affect parent nodes,but child node all feature of extends parent node!
// 子节点不能影响父节点，但子节点所有功能都继承自父节点!
func (zml *ZapMateLogger) SetAsyncer(chanLen uint) *ZapMateLogger {
	zml.lock.Lock()
	defer zml.lock.Unlock()
	if zml.isAsync {
		return zml
	}
	zml.isAsync = true
	zml.chanLen = chanLen
	zml.entryChan = make(chan *logEntry, zml.chanLen)
	logMsgPool = &sync.Pool{
		New: func() interface{} {
			return new(logEntry)
		},
	}
	go zml.startAsyncLogger()
	return zml
}

// startAsyncLogger 异步写入日志
func (zml *ZapMateLogger) startAsyncLogger() {
	zml.asyncWrite()
}

func (zml *ZapMateLogger) Debug(msg string, fields ...zap.Field) {
	zml.write(zap.DebugLevel, msg, fields...)
}

func (zml *ZapMateLogger) Info(msg string, fields ...zap.Field) {
	zml.write(zap.InfoLevel, msg, fields...)
}

func (zml *ZapMateLogger) Warn(msg string, fields ...zap.Field) {
	zml.write(zap.WarnLevel, msg, fields...)
}

func (zml *ZapMateLogger) Error(msg string, fields ...zap.Field) {
	zml.write(zap.ErrorLevel, msg, fields...)
}

func (zml *ZapMateLogger) DPanic(msg string, fields ...zap.Field) {
	zml.write(zap.DPanicLevel, msg, fields...)
}

func (zml *ZapMateLogger) Panic(msg string, fields ...zap.Field) {
	zml.write(zap.PanicLevel, msg, fields...)
}

func (zml *ZapMateLogger) Fatal(msg string, fields ...zap.Field) {
	zml.write(zap.FatalLevel, msg, fields...)
}

func (zml *ZapMateLogger) Flush() error {
	zml.flushWrite()
	zml.wg.Wait()
	return zml.Sync()
}
