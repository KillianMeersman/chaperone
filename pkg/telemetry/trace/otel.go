package trace

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var otelTracer trace.Tracer = noop.NewTracerProvider().Tracer("")

// Get a context with the necessary information for tracing, inherited from a request's headers.
func GetRequestContext(req *http.Request) context.Context {
	return otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
}

// Initialize tracing components.
// This includes the OpenTelemetry SDK, the OTLP exporter, and the tracer provider.
// It is recommended to call this function at the start of your application.
// The context should be cancelled when the application is shutting down to ensure proper cleanup.
func Init(ctx context.Context, res *resource.Resource) {
	exp, err := otlptracegrpc.New(ctx)
	if err != nil {
		panic(err)
	}

	// Create a new tracer provider with a batch span processor and the given exporter.
	// Ensure default SDK resources and the required service name are set.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	// Handle shutdown properly so nothing leaks.
	go func() {
		<-ctx.Done()
		tp.Shutdown(ctx)
	}()

	otel.SetTracerProvider(tp)

	// Finally, set the tracer that can be used for this package.
	otelTracer = tp.Tracer("")
}
