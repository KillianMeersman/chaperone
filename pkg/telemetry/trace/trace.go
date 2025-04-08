package trace

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Span struct {
	otelSpan trace.Span
}

// Record an event with the given attributes.
func (s *Span) Event(name string, attributes map[string]any) {
	otelAttributes := translateAttributes(attributes)
	s.otelSpan.AddEvent(name, trace.WithAttributes(otelAttributes...))
}

// Record an error that occured in a span.
func (s *Span) Error(err error) {
	s.otelSpan.RecordError(err)
}

func (s *Span) SetAttributes(attributes map[string]any) {
	s.otelSpan.SetAttributes(translateAttributes(attributes)...)
}

func (s *Span) End() {
	s.otelSpan.End()
}

func startSpan(ctx context.Context, kind trace.SpanKind, name string, attributes map[string]any) (context.Context, Span) {
	otelAttributes := translateAttributes(attributes)

	ctx, otelSpan := otelTracer.Start(ctx, name, trace.WithAttributes(otelAttributes...), trace.WithSpanKind(kind))
	return ctx, Span{
		otelSpan,
	}
}

func StartInternalSpan(ctx context.Context, name string, attributes map[string]any) (context.Context, Span) {
	return startSpan(ctx, trace.SpanKindInternal, name, attributes)
}

func StartServerSpan(ctx context.Context, name string, attributes map[string]any) (context.Context, Span) {
	return startSpan(ctx, trace.SpanKindServer, name, attributes)
}

func StartClientSpan(ctx context.Context, name string, attributes map[string]any) (context.Context, Span) {
	return startSpan(ctx, trace.SpanKindClient, name, attributes)
}

func StartConsumerSpan(ctx context.Context, name string, attributes map[string]any) (context.Context, Span) {
	return startSpan(ctx, trace.SpanKindConsumer, name, attributes)
}

func StartProducerSpan(ctx context.Context, name string, attributes map[string]any) (context.Context, Span) {
	return startSpan(ctx, trace.SpanKindProducer, name, attributes)
}

func CurrentSpan(ctx context.Context) Span {
	otelSpan := trace.SpanFromContext(ctx)
	return Span{
		otelSpan,
	}
}

func translateAttributes(attributes map[string]any) []attribute.KeyValue {
	otelAttributes := make([]attribute.KeyValue, 0, len(attributes))

	for k, v := range attributes {
		otelAttributes = append(otelAttributes, translateAttribute(k, v))
	}

	return otelAttributes
}

func translateAttribute(k string, v any) attribute.KeyValue {
	switch v := v.(type) {
	case string:
		return attribute.String(k, v)
	case int:
		return attribute.Int(k, v)
	case float64:
		return attribute.Float64(k, v)
	case bool:
		return attribute.Bool(k, v)
	case []string:
		return attribute.StringSlice(k, v)
	case []bool:
		return attribute.BoolSlice(k, v)
	}
	panic("unknown attribute type")
}
