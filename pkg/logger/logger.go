package logger

import (
	"sync"

	"go.uber.org/zap"
)

type ZapMateLogger struct {
	lock      *sync.Mutex
	isAsync   bool
	entryChan chan *logEntry
	chanLen   uint
	wg        *sync.WaitGroup
	*zap.Logger
}

func NewZapMateLogger(filename, section string) *ZapMateLogger {
	return &ZapMateLogger{
		isAsync:   false,
		lock:      new(sync.Mutex),
		Logger:    NewZapLogger(filename, section),
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
