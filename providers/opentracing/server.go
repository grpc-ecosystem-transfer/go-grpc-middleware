package opentracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/grpclog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/util/metautils"
)

func newServerSpanFromInbound(ctx context.Context, tracer opentracing.Tracer, traceHeaderName, fullMethodName string) (context.Context, opentracing.Span) {
	md := metautils.ExtractIncoming(ctx)
	parentSpanContext, err := tracer.Extract(opentracing.HTTPHeaders, metadataTextMap(md))
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		grpclog.Infof("grpc_opentracing: failed parsing trace information: %v", err)
	}

	serverSpan := tracer.StartSpan(
		fullMethodName,
		// this is magical, it attaches the new span to the parent parentSpanContext, and creates an unparented one if empty.
		ext.RPCServerOption(parentSpanContext),
		grpcTag,
	)

	// Log context information.
	t := tags.Extract(ctx)
	for k, v := range t.Values() {
		serverSpan.SetTag(k, v)
	}

	injectOpentracingIdsToTags(traceHeaderName, serverSpan, tags.Extract(ctx))
	return opentracing.ContextWithSpan(ctx, serverSpan), serverSpan
}
