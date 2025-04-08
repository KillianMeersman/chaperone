package telemetry

import (
	"context"

	"github.com/KillianMeersman/chaperone/pkg/telemetry/log"
	"github.com/KillianMeersman/chaperone/pkg/telemetry/trace"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const SERVICE_NAME = "chaperone"
const SERVICE_VERSION = "0.0.3"

// Initialize telemetry components.
// This includes tracing, logging, and metrics.
// It is recommended to call this function at the start of your application.
// The context should be cancelled when the application is shutting down to ensure proper cleanup.
func InitTelemetry(ctx context.Context, name, version string) {
	resource := getResource(name, version)
	log.Init(ctx, resource)
	trace.Init(ctx, resource)
}

func getResource(name, version string) *resource.Resource {
	r, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(name),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		panic(err)
	}
	return r
}
