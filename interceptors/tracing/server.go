// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package tracing

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/util/metautils"
)

var (
	grpcTag = opentracing.Tag{Key: string(ext.Component), Value: "gRPC"}
)

type serverReporter struct {
	ctx context.Context
	interceptors.CallMeta

	serverSpan opentracing.Span
}

func (o *serverReporter) PostCall(err error, _ time.Duration) {
	// Finish span and extract logging context information for richer spans.
	fieldIter := logging.ExtractFields(o.ctx).Iter()
	for fieldIter.Next() {
		o.serverSpan.SetTag(fieldIter.At())
	}
	if err != nil {
		ext.Error.Set(o.serverSpan, true)
		o.serverSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
	}
	o.serverSpan.Finish()
}

func (o *serverReporter) PostMsgSend(interface{}, error, time.Duration) {}

func (o *serverReporter) PostMsgReceive(interface{}, error, time.Duration) {}

type serverReportable struct {
	tracer          opentracing.Tracer
	traceHeaderName string
	filterOutFunc   FilterFunc
}

func (o *serverReportable) ServerReporter(ctx context.Context, c interceptors.CallMeta) (interceptors.Reporter, context.Context) {
	if o.filterOutFunc != nil && !o.filterOutFunc(ctx, c.FullMethod()) {
		return interceptors.NoopReporter{}, ctx
	}

	newCtx, serverSpan := newServerSpanFromInbound(ctx, o.tracer, o.traceHeaderName, c.FullMethod())
	mock := &serverReporter{
		ctx:        newCtx,
		CallMeta:   c,
		serverSpan: serverSpan,
	}
	return mock, newCtx
}

// UnaryServerInterceptor returns a new unary server interceptor for OpenTracing.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateOptions(opts)
	return interceptors.UnaryServerInterceptor(&serverReportable{tracer: o.tracer, traceHeaderName: o.traceHeaderName, filterOutFunc: o.filterOutFunc})
}

// StreamServerInterceptor returns a new streaming server interceptor for OpenTracing.
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateOptions(opts)
	return interceptors.StreamServerInterceptor(&serverReportable{tracer: o.tracer, traceHeaderName: o.traceHeaderName, filterOutFunc: o.filterOutFunc})
}

func newServerSpanFromInbound(ctx context.Context, tracer opentracing.Tracer, traceHeaderName, fullMethodName string) (context.Context, opentracing.Span) {
	md := metautils.ExtractIncoming(ctx)
	parentSpanContext, err := tracer.Extract(opentracing.HTTPHeaders, metadataTextMap(md))
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		grpclog.Infof("grpc_opentracing: failed parsing trace information: %v", err)
	}

	serverSpan := tracer.StartSpan(
		fullMethodName,
		// This is magical, it attaches the new span to the parent parentSpanContext, and creates an unparented one if empty.
		ext.RPCServerOption(parentSpanContext),
		grpcTag,
	)

	meta := getTraceMeta(traceHeaderName, serverSpan)

	// Logging fields are used as input for span finish tags. We also want request/trace ID to be part of logging.
	// Use logging fields to preserve this information.
	ctx = logging.InjectFields(ctx, logging.Fields{FieldTraceID, meta.TraceID, FieldSpanID, meta.SpanID, FieldSampled, fmt.Sprintf("%v", meta.Sampled)})
	return opentracing.ContextWithSpan(ctx, serverSpan), serverSpan
}
