package log

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/log/noop"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

var otelLogger = noop.NewLoggerProvider().Logger("")

// Initialize OpenTelemetry logging components.
// The log package is automatically instrumented to use OpenTelemetry.
// This includes the OpenTelemetry SDK, the OTLP exporter, and the logger provider.
// It is recommended to call this function at the start of your application.
// The context should be cancelled when the application is shutting down to ensure proper cleanup.
func Init(ctx context.Context, res *resource.Resource) {
	// Create a logger provider.
	// You can pass this instance directly when creating bridges.
	exporter, err := otlploggrpc.New(ctx)
	if err != nil {
		panic(err)
	}
	processor := otellog.NewBatchProcessor(exporter)
	provider := otellog.NewLoggerProvider(
		otellog.WithResource(res),
		otellog.WithProcessor(processor),
	)

	// Handle shutdown properly so nothing leaks.
	go func() {
		<-ctx.Done()
		Info(ctx, "shutting down logger provider")
		if err := provider.Shutdown(ctx); err != nil {
			Error(ctx, err)
		}
	}()

	// Register as global logger provider so that it can be accessed global.LoggerProvider.
	// Most log bridges use the global logger provider as default.
	// If the global logger provider is not set then a no-op implementation
	// is used, which fails to generate data.
	global.SetLoggerProvider(provider)
	otelLogger = provider.Logger("")
}
