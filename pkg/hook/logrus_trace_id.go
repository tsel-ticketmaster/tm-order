package hook

import (
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type traceIDLoggerHook struct {
}

func NewTraceIDLoggerHook() logrus.Hook {

	return &traceIDLoggerHook{}
}

// Levels implements logrus.Hook interface, this hook applies to all defined levels
func (h *traceIDLoggerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire implements logrus.Hook interface, attaches trace and span details found in entry context
func (h *traceIDLoggerHook) Fire(e *logrus.Entry) error {
	ctx := e.Context
	if ctx == nil {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	if span == nil {
		return nil
	}

	traceID := span.SpanContext().TraceID().String()

	e.Data["trace_id"] = traceID

	return nil
}
