package log

import (
	"github.com/rigado/ble"
	"go.uber.org/zap"
)

func NewBLELogger(logger *zap.Logger) ble.Logger {
	return &bleLogger{logger: logger.WithOptions(zap.WithCaller(false)).Sugar()}
}

type bleLogger struct {
	logger *zap.SugaredLogger
}

func (b *bleLogger) Info(v ...any)               { b.logger.Info(v...) }
func (b *bleLogger) Debug(v ...any)              { b.logger.Debug(v...) }
func (b *bleLogger) Error(v ...any)              { b.logger.Error(v...) }
func (b *bleLogger) Warn(v ...any)               { b.logger.Warn(v...) }
func (b *bleLogger) Infof(msg string, v ...any)  { b.logger.Infof(msg, v...) }
func (b *bleLogger) Debugf(msg string, v ...any) { b.logger.Infof(msg, v...) }
func (b *bleLogger) Errorf(msg string, v ...any) { b.logger.Errorf(msg, v...) }
func (b *bleLogger) Warnf(msg string, v ...any)  { b.logger.Errorf(msg, v...) }

func (b *bleLogger) ChildLogger(tags map[string]any) ble.Logger {
	opts := make([]zap.Field, 0, len(tags))

	for k, v := range tags {
		opts = append(opts, zap.Any(k, v))
	}

	return &bleLogger{logger: b.logger.Desugar().With(opts...).Sugar()}
}
