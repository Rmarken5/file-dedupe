package infra

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

func InitTracer() (*trace.TracerProvider, error) {
	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint("localhost:4318"))
	if err != nil {
		return nil, fmt.Errorf("cannot create new tracing exporter: %w", err)
	}

	traceProvider := trace.NewTracerProvider(trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("file dedupe"))))
	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}
