package xlogger

import (
	"context"

	"github.com/sirupsen/logrus"
)

type ILogger interface {
	InfoF(ctx context.Context, format string, args ...interface{})
	PanicF(ctx context.Context, format string, args ...interface{})
	ErrorF(ctx context.Context, format string, args ...interface{})
	WarnF(ctx context.Context, format string, args ...interface{})
	FatalF(ctx context.Context, format string, args ...interface{})
}

var Logger ILogger = logrusLog{}

type logrusLog struct {
}

func (l logrusLog) InfoF(ctx context.Context, format string, args ...interface{}) {
	logrus.WithContext(ctx).Infof(format, args...)
}

func (l logrusLog) PanicF(ctx context.Context, format string, args ...interface{}) {
	logrus.WithContext(ctx).Panicf(format, args...)
}

func (l logrusLog) ErrorF(ctx context.Context, format string, args ...interface{}) {
	logrus.WithContext(ctx).Errorf(format, args...)
}

func (l logrusLog) WarnF(ctx context.Context, format string, args ...interface{}) {
	logrus.WithContext(ctx).Warnf(format, args...)
}

func (l logrusLog) FatalF(ctx context.Context, format string, args ...interface{}) {
	logrus.WithContext(ctx).Fatalf(format, args...)
}

func InfoF(ctx context.Context, format string, args ...interface{}) {
	Logger.InfoF(ctx, format, args...)
}

func PanicF(ctx context.Context, format string, args ...interface{}) {
	Logger.PanicF(ctx, format, args...)
}

func ErrorF(ctx context.Context, format string, args ...interface{}) {
	Logger.ErrorF(ctx, format, args...)
}

func WarnF(ctx context.Context, format string, args ...interface{}) {
	Logger.WarnF(ctx, format, args...)
}

func FatalF(ctx context.Context, format string, args ...interface{}) {
	Logger.FatalF(ctx, format, args...)
}
