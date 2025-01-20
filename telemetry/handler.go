package telemetry

import (
	"context"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	t "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/yeldiRium/hledger-language-server/version"
)

type Telemetry struct {
	logger *zap.Logger
	Tracer t.Tracer
}

func SetupTelemetry(ctx context.Context, logger *zap.Logger) (*Telemetry, func(context.Context) error, error) {
	resource, err := resource.New(ctx, resource.WithAttributes(
		attribute.String("service.name", "hledger-language-server"),
	))
	if err != nil {
		return nil, nil, err
	}

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)

	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithResource(resource),
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(5*time.Second)),
	)

	otel.SetTracerProvider(traceProvider)
	tracer := traceProvider.Tracer(version.Path())

	return &Telemetry{
		logger: logger,
		Tracer: tracer,
	}, traceProvider.Shutdown, nil
}

func SetGlobalAttributes(span t.Span) {
	span.SetAttributes(
		attribute.String("service.name", "hledger-language-server"),
		attribute.String("service.version", version.Version()),
		attribute.String("service.build.checksum", version.Sum()),
		attribute.String("service.build.git_hash", version.CommitHash()),
		attribute.String("service.build.git_time", version.CommitTime()),
		attribute.String("service.build.git_dirty", version.Dirty()),
	)
}

func WrapInTelemetry(telemetry *Telemetry, innerHandler jsonrpc2.Handler) jsonrpc2.Handler {
	wrappedHandler := func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		ctx, span := telemetry.Tracer.Start(ctx, req.Method())
		defer span.End()

		SetGlobalAttributes(span)

		err := innerHandler(ctx, reply, req)
		if err != nil {
			span.SetStatus(codes.Error, "handler failed")
			span.RecordError(err)
			return err
		}

		return nil
	}

	return wrappedHandler
}
