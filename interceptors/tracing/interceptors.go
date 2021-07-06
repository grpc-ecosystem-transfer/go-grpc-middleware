package tracing

import (
	"context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

type SpanKind string

const (
	SpanKindServer SpanKind = "server"
	SpanKindClient SpanKind = "client"
)

type reportable struct {
	tracer Tracer
}

func (r *reportable) ServerReporter(ctx context.Context, _ interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	return r.reporter(ctx, service, method, SpanKindServer)
}

func (r *reportable) ClientReporter(ctx context.Context, _ interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	return r.reporter(ctx, service, method, SpanKindClient)
}

func (r *reportable) reporter(ctx context.Context, service string, method string, kind SpanKind) (interceptors.Reporter, context.Context) {
	newCtx, span := r.tracer.Start(ctx, interceptors.FullMethod(service, method), kind)
	reporter := reporter{ctx: newCtx, span: span}

	return &reporter, newCtx
}

// UnaryClientInterceptor returns a new unary client interceptor that optionally traces the execution of external gRPC calls.
// Tracer will use tags (from tags package) available in current context as fields.
func UnaryClientInterceptor(tracer Tracer) grpc.UnaryClientInterceptor {
	return interceptors.UnaryClientInterceptor(&reportable{tracer: tracer})
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally traces the execution of external gRPC calls.
// Tracer will use tags (from tags package) available in current context as fields.
func StreamClientInterceptor(tracer Tracer) grpc.StreamClientInterceptor {
	return interceptors.StreamClientInterceptor(&reportable{tracer: tracer})
}

// UnaryServerInterceptor returns a new unary server interceptors that optionally traces endpoint handling.
// Tracer will use tags (from tags package) available in current context as fields.
func UnaryServerInterceptor(tracer Tracer) grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(&reportable{tracer: tracer})
}

// StreamServerInterceptor returns a new stream server interceptors that optionally traces endpoint handling.
// Tracer will use tags (from tags package) available in current context as fields.
func StreamServerInterceptor(tracer Tracer) grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(&reportable{tracer: tracer})
}
