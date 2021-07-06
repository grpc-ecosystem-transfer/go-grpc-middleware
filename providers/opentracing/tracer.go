package opentracing

import (
	"context"

	"github.com/opentracing/opentracing-go"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
)

type tracer struct {
	tracer opentracing.Tracer
	// This is only used for server.
	traceHeaderName string
}

// Compatibility check.
var _ tracing.Tracer = &tracer{}

// InterceptorTracer converts OpenTracing tracer to Tracer adapter.
func InterceptorTracer(opts ...Option) *tracer {
	o := evaluateOptions(opts)

	return &tracer{tracer: o.tracer, traceHeaderName: o.traceHeaderName}
}

func (t *tracer) Start(ctx context.Context, spanName string, kind tracing.SpanKind) (context.Context, tracing.Span) {
	var span opentracing.Span
	switch kind {
	case tracing.SpanKindClient:
		ctx, span = newClientSpanFromContext(ctx, t.tracer, spanName)
	case tracing.SpanKindServer:
		ctx, span = newServerSpanFromInbound(ctx, t.tracer, t.traceHeaderName, spanName)
	}
	return ctx, newSpan(span, ctx)
}
