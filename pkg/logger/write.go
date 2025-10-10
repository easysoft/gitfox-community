package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// write 写入日志
func (zml *ZapMateLogger) write(lvl zapcore.Level, msg string, fields ...zap.Field) {
	// The time of the log is generated here! 日志的时间在这里生成！
	if ce := zml.Logger.Check(lvl, msg); ce != nil {
		// if isAsync so... 异步写入日志
		if zml.isAsync {
			le := logMsgPool.Get().(*logEntry)
			le.entry = ce
			le.fields = fields
			zml.wg.Add(1)
			zml.entryChan <- le
		} else { //if isSync so... 同步写入日志
			ce.Write(fields...)
		}
	}
}

// asyncWrite 异步写入日志
func (zml *ZapMateLogger) asyncWrite() {
	for {
		select {
		case le := <-zml.entryChan:
			le.entry.Write(le.fields...)
			// There must used Sync.Put,to avoid a lot of GC 必须使用Sync.Put，以避免大量的GC
			logMsgPool.Put(le) //
			zml.wg.Done()
		}
	}
}

// Flush 刷新异步写入的日志
func (zml *ZapMateLogger) flushWrite() {
	for {
		select {
		case le := <-zml.entryChan:
			le.entry.Write(le.fields...)
			// There must used Sync.Put,to avoid a lot of GC 必须使用Sync.Put，以避免大量的GC
			logMsgPool.Put(le)
			zml.wg.Done()
		default: // if there is no data in the channel,so finsh sync flush and exit  如果通道中没有数据，那么就完成同步刷新并退出
			return
		}
	}
}
