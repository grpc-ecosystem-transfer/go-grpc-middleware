package opentelemetry

import (
	"context"
	
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
)

type tracer struct {
	tracer      trace.Tracer
	propagators propagation.TextMapPropagator
}

// Compatibility check.
var _ tracing.Tracer = &tracer{}

// InterceptorTracer converts OpenTelemetry tracer to Tracer adapter.
func InterceptorTracer(opts ...Option) *tracer {
	o := newConfig(opts...)
	
	return &tracer{tracer: o.tracer, propagators: o.Propagators}
}

func (t *tracer) Start(ctx context.Context, spanName string, kind tracing.SpanKind) (context.Context, tracing.Span) {
	var span trace.Span
	switch kind {
	case tracing.SpanKindClient:
		ctx, span = newClientSpanFromContext(ctx, spanName, t.tracer, t.propagators)
	case tracing.SpanKindServer:
		ctx, span = newServerSpanFromContext(ctx, spanName, t.tracer, t.propagators)
	}
	return ctx, newSpan(ctx, span)
}
