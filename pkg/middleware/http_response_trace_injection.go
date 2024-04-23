package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

func HTTPResponseTraceInjection(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer handler.ServeHTTP(w, r)

		span := trace.SpanFromContext(r.Context())
		if span == nil {
			return
		}

		traceID := span.SpanContext().TraceID().String()
		w.Header().Set("X-TRACE-ID", traceID)
	})
}
