package logger

import (
	"fmt"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Note: base.Logger and sugar.Logger must be same pointer
// 注: base.Logger和sugar.Logger必须是相同的指针
type SugaredMateLogger struct {
	base *ZapMateLogger
	*zap.SugaredLogger
}

// Sugar 返回一个 SugaredLogger
// 注: SugaredLogger 是 Logger 的 sugared 版本，它提供了更方便的日志记录方法
func (log *ZapMateLogger) Sugar() *SugaredMateLogger {
	return &SugaredMateLogger{
		log.clone(),
		log.Logger.Sugar(),
	}
}

// Desugar 返回一个 Logger
// 注: Logger 是 SugaredLogger 的 desugared 版本，它提供了更底层的日志记录方法
func (s *SugaredMateLogger) Desugar() *ZapMateLogger {
	base := s.base.clone()
	base.Logger = s.SugaredLogger.Desugar()
	return base
}

func (s *SugaredMateLogger) Named(name string) *SugaredMateLogger {
	return &SugaredMateLogger{
		base:          s.base.Named(name),
		SugaredLogger: s.SugaredLogger.Named(name),
	}
}

func (s *SugaredMateLogger) With(args ...interface{}) *SugaredMateLogger {
	return &SugaredMateLogger{
		base:          s.base.With(s.sweetenFields(args)...),
		SugaredLogger: s.SugaredLogger.With(args...),
	}
}

func (s *SugaredMateLogger) Debug(args ...interface{}) {
	s.asynclog(zap.DebugLevel, "", args, nil)
}

func (s *SugaredMateLogger) Info(args ...interface{}) {
	s.asynclog(zap.InfoLevel, "", args, nil)
}

func (s *SugaredMateLogger) Warn(args ...interface{}) {
	s.asynclog(zap.WarnLevel, "", args, nil)
}

func (s *SugaredMateLogger) Error(args ...interface{}) {
	s.asynclog(zap.ErrorLevel, "", args, nil)
}

func (s *SugaredMateLogger) DPanic(args ...interface{}) {
	s.asynclog(zap.DPanicLevel, "", args, nil)
}

func (s *SugaredMateLogger) Panic(args ...interface{}) {
	s.asynclog(zap.PanicLevel, "", args, nil)
}

func (s *SugaredMateLogger) Fatal(args ...interface{}) {
	s.asynclog(zap.FatalLevel, "", args, nil)
}

func (s *SugaredMateLogger) Debugf(template string, args ...interface{}) {
	s.asynclog(zap.DebugLevel, template, args, nil)
}

func (s *SugaredMateLogger) Infof(template string, args ...interface{}) {
	s.asynclog(zap.InfoLevel, template, args, nil)
}

func (s *SugaredMateLogger) Warnf(template string, args ...interface{}) {
	s.asynclog(zap.WarnLevel, template, args, nil)
}

func (s *SugaredMateLogger) Errorf(template string, args ...interface{}) {
	s.asynclog(zap.ErrorLevel, template, args, nil)
}

func (s *SugaredMateLogger) DPanicf(template string, args ...interface{}) {
	s.asynclog(zap.DPanicLevel, template, args, nil)
}

func (s *SugaredMateLogger) Panicf(template string, args ...interface{}) {
	s.asynclog(zap.PanicLevel, template, args, nil)
}

func (s *SugaredMateLogger) Fatalf(template string, args ...interface{}) {
	s.asynclog(zap.FatalLevel, template, args, nil)
}

func (s *SugaredMateLogger) Debugw(msg string, keysAndValues ...interface{}) {
	s.asynclog(zap.DebugLevel, msg, nil, keysAndValues)
}

func (s *SugaredMateLogger) Infow(msg string, keysAndValues ...interface{}) {
	s.asynclog(zap.InfoLevel, msg, nil, keysAndValues)
}

func (s *SugaredMateLogger) Warnw(msg string, keysAndValues ...interface{}) {
	s.asynclog(zap.WarnLevel, msg, nil, keysAndValues)
}

func (s *SugaredMateLogger) Errorw(msg string, keysAndValues ...interface{}) {
	s.asynclog(zap.ErrorLevel, msg, nil, keysAndValues)
}

func (s *SugaredMateLogger) DPanicw(msg string, keysAndValues ...interface{}) {
	s.asynclog(zap.DPanicLevel, msg, nil, keysAndValues)
}

func (s *SugaredMateLogger) Panicw(msg string, keysAndValues ...interface{}) {
	s.asynclog(zap.PanicLevel, msg, nil, keysAndValues)
}

func (s *SugaredMateLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	s.asynclog(zap.FatalLevel, msg, nil, keysAndValues)
}

func (s *SugaredMateLogger) Flush() error {
	return s.base.Flush()
}

// ---- 实现了 Sugar Logger（语法糖日志器） 的核心功能 ----

const (
	_oddNumberErrMsg    = "Ignored key without a value."
	_nonStringKeyErrMsg = "Ignored key-value pairs with non-string keys."
)

func (s *SugaredMateLogger) asynclog(lvl zapcore.Level, template string, fmtArgs []interface{}, context []interface{}) {

	if lvl < zap.DPanicLevel && !s.base.Core().Enabled(lvl) {
		return
	}

	msg := template
	if msg == "" && len(fmtArgs) > 0 {
		msg = fmt.Sprint(fmtArgs...)
	} else if msg != "" && len(fmtArgs) > 0 {
		msg = fmt.Sprintf(template, fmtArgs...)
	}

	s.base.write(lvl, msg, s.sweetenFields(context)...)
}

func (s *SugaredMateLogger) sweetenFields(args []interface{}) []zap.Field {
	if len(args) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(args))
	var invalid invalidPairs

	for i := 0; i < len(args); {
		if f, ok := args[i].(zap.Field); ok {
			fields = append(fields, f)
			i++
			continue
		}

		if i == len(args)-1 {
			s.base.DPanic(_oddNumberErrMsg, zap.Any("ignored", args[i]))
			break
		}

		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); !ok {
			if cap(invalid) == 0 {
				invalid = make(invalidPairs, 0, len(args)/2)
			}
			invalid = append(invalid, invalidPair{i, key, val})
		} else {
			fields = append(fields, zap.Any(keyStr, val))
		}
		i += 2
	}

	if len(invalid) > 0 {
		s.base.DPanic(_nonStringKeyErrMsg, zap.Array("invalid", invalid))
	}
	return fields
}

type invalidPair struct {
	position   int
	key, value interface{}
}

type invalidPairs []invalidPair

func (p invalidPair) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("position", int64(p.position))
	zap.Any("key", p.key).AddTo(enc)
	zap.Any("value", p.value).AddTo(enc)
	return nil
}

func (ps invalidPairs) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	var err error
	for i := range ps {
		err = multierr.Append(err, enc.AppendObject(ps[i]))
	}
	return err
}
