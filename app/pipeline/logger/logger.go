package logger

import (
	"context"
	"fmt"

	dronelogger "github.com/drone/runner-go/logger"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

// WithWrappedLogger adds a wrapped copy of the app logger to the context,
// satisfying the drone/runner-go logger interface.
func WithWrappedLogger(ctx context.Context) context.Context {
	return dronelogger.WithContext(ctx, &wrapLogger{
		inner: logger.GetLogger(ctx),
		err:   nil,
	})
}

// WithUnwrappedLogger extracts the wrapped app logger from the context and
// stores it back as a plain app logger for downstream code.
func WithUnwrappedLogger(ctx context.Context) context.Context {
	if wrapped, ok := dronelogger.FromContext(ctx).(*wrapLogger); ok {
		return logger.WithLogger(ctx, wrapped.inner)
	}
	return logger.WithLogger(ctx, logger.GetLogger(ctx))
}

// wrapLogger implements the drone/runner-go dronelogger.Logger interface on top of
// the project's SugaredMateLogger.
type wrapLogger struct {
	inner *logger.SugaredMateLogger
	err   error
}

func (w *wrapLogger) WithError(err error) dronelogger.Logger {
	return &wrapLogger{inner: w.inner, err: err}
}

func (w *wrapLogger) WithField(key string, value interface{}) dronelogger.Logger {
	return &wrapLogger{inner: w.inner.With(key, value), err: w.err}
}

func (w *wrapLogger) Debug(args ...interface{}) {
	w.write(w.inner.Debug, w.inner.Debugw, args)
}

func (w *wrapLogger) Debugf(format string, args ...interface{}) {
	if w.err != nil {
		w.inner.Debugw(fmt.Sprintf(format, args...), "error", w.err)
		return
	}
	w.inner.Debugf(format, args...)
}

func (w *wrapLogger) Debugln(args ...interface{}) {
	w.write(w.inner.Debug, w.inner.Debugw, args)
}

func (w *wrapLogger) Error(args ...interface{}) {
	w.write(w.inner.Error, w.inner.Errorw, args)
}

func (w *wrapLogger) Errorf(format string, args ...interface{}) {
	if w.err != nil {
		w.inner.Errorw(fmt.Sprintf(format, args...), "error", w.err)
		return
	}
	w.inner.Errorf(format, args...)
}

func (w *wrapLogger) Errorln(args ...interface{}) {
	w.write(w.inner.Error, w.inner.Errorw, args)
}

func (w *wrapLogger) Info(args ...interface{}) {
	w.write(w.inner.Info, w.inner.Infow, args)
}

func (w *wrapLogger) Infof(format string, args ...interface{}) {
	if w.err != nil {
		w.inner.Infow(fmt.Sprintf(format, args...), "error", w.err)
		return
	}
	w.inner.Infof(format, args...)
}

func (w *wrapLogger) Infoln(args ...interface{}) {
	w.write(w.inner.Info, w.inner.Infow, args)
}

func (w *wrapLogger) Trace(args ...interface{}) {
	w.write(w.inner.Trace, w.inner.Debugw, args)
}

func (w *wrapLogger) Tracef(format string, args ...interface{}) {
	if w.err != nil {
		w.inner.Debugw(fmt.Sprintf(format, args...), "error", w.err)
		return
	}
	w.inner.Tracef(format, args...)
}

func (w *wrapLogger) Traceln(args ...interface{}) {
	w.write(w.inner.Trace, w.inner.Debugw, args)
}

func (w *wrapLogger) Warn(args ...interface{}) {
	w.write(w.inner.Warn, w.inner.Warnw, args)
}

func (w *wrapLogger) Warnf(format string, args ...interface{}) {
	if w.err != nil {
		w.inner.Warnw(fmt.Sprintf(format, args...), "error", w.err)
		return
	}
	w.inner.Warnf(format, args...)
}

func (w *wrapLogger) Warnln(args ...interface{}) {
	w.write(w.inner.Warn, w.inner.Warnw, args)
}

// write 根据是否携带 error 选择普通或带结构化字段的日志方法
func (w *wrapLogger) write(plain func(...interface{}), withErr func(string, ...interface{}), args []interface{}) {
	if w.err != nil {
		withErr(fmt.Sprint(args...), "error", w.err)
		return
	}
	plain(fmt.Sprint(args...))
}
