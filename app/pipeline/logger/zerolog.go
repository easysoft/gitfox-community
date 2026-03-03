package logger

import (
	"context"
	"fmt"

	"github.com/drone/runner-go/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// WithWrappedZerolog adds a wrapped copy of the zerolog logger to the context.
func WithWrappedZerolog(ctx context.Context) context.Context {
	return logger.WithContext(
		ctx,
		&wrapZerolog{
			inner: log.Ctx(ctx).With().Logger(),
			err:   nil,
		})
}

// WithUnwrappedZerolog adds an unwrapped copy of the zerolog logger to the context.
func WithUnwrappedZerolog(ctx context.Context) context.Context {
	// try to get the logger from the wrapped zerologger
	if wrappedLogger, ok := logger.FromContext(ctx).(*wrapZerolog); ok {
		return wrappedLogger.inner.WithContext(ctx)
	}

	// if there's no logger, fall-back to global logger instance
	return log.Logger.WithContext(ctx)
}

// wrapZerolog wraps the zerolog logger to be used within drone packages.
type wrapZerolog struct {
	inner zerolog.Logger
	err   error
}

func (w *wrapZerolog) WithError(err error) logger.Logger {
	return &wrapZerolog{inner: w.inner, err: err}
}

func (w *wrapZerolog) WithField(key string, value interface{}) logger.Logger {
	return &wrapZerolog{inner: w.inner.With().Str(key, fmt.Sprint(value)).Logger(), err: w.err}
}

func (w *wrapZerolog) Debug(args ...interface{}) {
	w.inner.Debug().Err(w.err).Msg(fmt.Sprint(args...))
}

func (w *wrapZerolog) Debugf(format string, args ...interface{}) {
	w.inner.Debug().Err(w.err).Msgf(format, args...)
}

func (w *wrapZerolog) Debugln(args ...interface{}) {
	w.inner.Debug().Err(w.err).Msg(fmt.Sprintln(args...))
}

func (w *wrapZerolog) Error(args ...interface{}) {
	w.inner.Error().Err(w.err).Msg(fmt.Sprint(args...))
}

func (w *wrapZerolog) Errorf(format string, args ...interface{}) {
	w.inner.Error().Err(w.err).Msgf(format, args...)
}

func (w *wrapZerolog) Errorln(args ...interface{}) {
	w.inner.Error().Err(w.err).Msg(fmt.Sprintln(args...))
}

func (w *wrapZerolog) Info(args ...interface{}) {
	w.inner.Info().Err(w.err).Msg(fmt.Sprint(args...))
}

func (w *wrapZerolog) Infof(format string, args ...interface{}) {
	w.inner.Info().Err(w.err).Msgf(format, args...)
}

func (w *wrapZerolog) Infoln(args ...interface{}) {
	w.inner.Info().Err(w.err).Msg(fmt.Sprintln(args...))
}

func (w *wrapZerolog) Trace(args ...interface{}) {
	w.inner.Trace().Err(w.err).Msg(fmt.Sprint(args...))
}

func (w *wrapZerolog) Tracef(format string, args ...interface{}) {
	w.inner.Trace().Err(w.err).Msgf(format, args...)
}

func (w *wrapZerolog) Traceln(args ...interface{}) {
	w.inner.Trace().Err(w.err).Msg(fmt.Sprintln(args...))
}

func (w *wrapZerolog) Warn(args ...interface{}) {
	w.inner.Warn().Err(w.err).Msg(fmt.Sprint(args...))
}

func (w *wrapZerolog) Warnf(format string, args ...interface{}) {
	w.inner.Warn().Err(w.err).Msgf(format, args...)
}

func (w *wrapZerolog) Warnln(args ...interface{}) {
	w.inner.Warn().Err(w.err).Msg(fmt.Sprintln(args...))
}
