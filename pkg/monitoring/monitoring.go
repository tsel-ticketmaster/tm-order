package monitoring

import (
	"context"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type OpenTelemetry interface {
	Start(ctx context.Context) (err error)
	Stop(ctx context.Context) (err error)
}

type openTelemetryImpl struct {
	serviceName   string
	environment   string
	gcpProjectID  string
	resource      *resource.Resource
	traceExporter *texporter.Exporter
}

func NewOpenTelemetry(serviceName, environment, gcpProjectID string) OpenTelemetry {
	ctx := context.Background()
	res, _ := resource.New(
		ctx,
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("environment", environment),
		),
	)

	return &openTelemetryImpl{
		serviceName:  serviceName,
		environment:  environment,
		gcpProjectID: gcpProjectID,
		resource:     res,
	}
}

// Start implements OpenTelemetry.
func (ot *openTelemetryImpl) Start(ctx context.Context) (err error) {
	exporter, err := texporter.New(texporter.WithProjectID(ot.gcpProjectID))
	if err != nil {
		otel.Handle(err)
		return
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(ot.resource),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(provider)

	ot.traceExporter = exporter

	return
}

// Stop implements OpenTelemetry.
func (ot *openTelemetryImpl) Stop(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if ot.traceExporter == nil {
		return nil
	}

	err = ot.traceExporter.Shutdown(ctx)
	if err != nil {
		otel.Handle(err)
		return
	}

	return
}
