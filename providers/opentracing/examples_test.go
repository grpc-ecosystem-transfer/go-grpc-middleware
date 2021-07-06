package opentracing_test

import (
	"testing"
	
	"google.golang.org/grpc"
	
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	grpcopentracing "github.com/grpc-ecosystem/go-grpc-middleware/v2/providers/opentracing"
)

func Example() {
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			tags.UnaryServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			tracing.UnaryServerInterceptor(grpcopentracing.InterceptorTracer()),
		),
		grpc.ChainStreamInterceptor(
			tags.StreamServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			tracing.StreamServerInterceptor(grpcopentracing.InterceptorTracer()),
		),
	)
	
	_, _ = grpc.Dial("",
		grpc.WithUnaryInterceptor(tracing.UnaryClientInterceptor(grpcopentracing.InterceptorTracer())),
		grpc.WithStreamInterceptor(tracing.StreamClientInterceptor(grpcopentracing.InterceptorTracer())),
	)
}

func TestExamplesBuildable(t *testing.T) {
	Example()
}
