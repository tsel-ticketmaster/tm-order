package applogger

import (
	"context"
	"fmt"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type ZapLogger struct {
	l *otelzap.Logger
}

func GetZapLogger() *ZapLogger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.LevelKey = "severity"

	zapLogger, err := cfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(
			zap.String("app.type", "be-web"),
			zap.String("app.name", "user"),
		),
	)

	fmt.Println("zap error:", err)

	logger := otelzap.New(zapLogger)

	return &ZapLogger{l: logger}
}

func (z *ZapLogger) Error(ctx context.Context, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	z.l.Ctx(ctx).Error(err.Error(), fields...)
}

func (z *ZapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	z.l.Ctx(ctx).Info(msg, fields...)
}
