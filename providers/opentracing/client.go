package opentracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/grpclog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/util/metautils"
)

var (
	grpcTag = opentracing.Tag{Key: string(ext.Component), Value: "gRPC"}
)

// ClientAddContextTags returns a context with specified opentracing tags, which
// are used by UnaryClientInterceptor/StreamClientInterceptor when creating a
// new span.
func ClientAddContextTags(ctx context.Context, tags opentracing.Tags) context.Context {
	return context.WithValue(ctx, clientSpanTagKey{}, tags)
}

type clientSpanTagKey struct{}

func newClientSpanFromContext(ctx context.Context, tracer opentracing.Tracer, fullMethodName string) (context.Context, opentracing.Span) {
	var parentSpanCtx opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentSpanCtx = parent.Context()
	}
	opts := []opentracing.StartSpanOption{
		opentracing.ChildOf(parentSpanCtx),
		ext.SpanKindRPCClient,
		grpcTag,
	}
	if tagx := ctx.Value(clientSpanTagKey{}); tagx != nil {
		if opt, ok := tagx.(opentracing.StartSpanOption); ok {
			opts = append(opts, opt)
		}
	}
	clientSpan := tracer.StartSpan(fullMethodName, opts...)
	// Make sure we add this to the metadata of the call, so it gets propagated:
	md := metautils.ExtractOutgoing(ctx).Clone()
	if err := tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, metadataTextMap(md)); err != nil {
		grpclog.Infof("grpc_opentracing: failed serializing trace information: %v", err)
	}
	ctxWithMetadata := md.ToOutgoing(ctx)
	return opentracing.ContextWithSpan(ctxWithMetadata, clientSpan), clientSpan
}
