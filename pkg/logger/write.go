// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

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
	for le := range zml.entryChan {
		// 写入日志
		le.entry.Write(le.fields...)
		// 清理 fields 切片，避免内存泄漏
		le.fields = nil
		// 将 logEntry 放回对象池，减少 GC 压力
		logMsgPool.Put(le)
		// 标记此日志条目已处理完成
		zml.wg.Done()
	}
}

// flushWrite 刷新异步写入的日志
func (zml *ZapMateLogger) flushWrite() {
	for {
		select {
		case le, ok := <-zml.entryChan:
			if !ok {
				// Channel 已关闭，退出
				return
			}
			// 写入日志
			le.entry.Write(le.fields...)
			// 清理 fields 切片，避免内存泄漏
			le.fields = nil
			// 将 logEntry 放回对象池，减少 GC 压力
			logMsgPool.Put(le)
			// 标记此日志条目已处理完成
			zml.wg.Done()
		default:
			// 如果通道中没有数据，完成同步刷新并退出
			return
		}
	}
}
