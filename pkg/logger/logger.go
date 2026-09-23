// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package logger

import (
	"sync"

	"go.uber.org/zap"

	"github.com/easysoft/gitfox-community/pkg/config"
)

type ZapMateLogger struct {
	lock      *sync.Mutex
	isAsync   bool
	entryChan chan *logEntry
	chanLen   uint
	wg        *sync.WaitGroup
	*zap.Logger
}

func NewZapMateLogger(config config.LoggingConfig, logType string) *ZapMateLogger {
	return &ZapMateLogger{
		isAsync:   false,
		lock:      new(sync.Mutex),
		Logger:    NewZapLogger(config, logType),
		entryChan: make(chan *logEntry, 1),
		wg:        new(sync.WaitGroup),
	}
}

func (log *ZapMateLogger) With(fields ...zap.Field) *ZapMateLogger {
	copy := log.clone()
	copy.Logger = log.Logger.With(fields...)
	return copy
}

func (log *ZapMateLogger) Named(s string) *ZapMateLogger {
	copy := log.clone()
	copy.Logger = log.Logger.Named(s)
	return copy
}

func (log *ZapMateLogger) clone() *ZapMateLogger {
	copy := *log
	return &copy
}
