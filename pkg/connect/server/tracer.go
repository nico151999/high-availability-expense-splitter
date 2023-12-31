package server

import (
	"context"

	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func createOtlpExporter(ctx context.Context, collectorUrl string) (sdktrace.SpanExporter, error) {
	log := logging.FromContext(ctx)

	log.Debug("creating OpenTelemetry exporter")
	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(collectorUrl),
			// we assume service meshes to apply mTLS for us
			otlptracegrpc.WithInsecure(),
		),
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed creating exporter")
	}
	return exporter, nil
}

// initTracer configures a trace provider based on an otlp trace exporter
func initTracer(ctx context.Context, serviceName string, exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	log := logging.FromContext(ctx)

	log.Debug("creating resources for trace provider")
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		return nil, eris.Wrap(err, "could not create resources for trace provider")
	}

	log.Debug("creating trace provider")
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return traceProvider, nil
}
