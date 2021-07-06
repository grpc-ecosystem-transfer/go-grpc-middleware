package opentelemetry_test

import (
	"testing"
	
	"google.golang.org/grpc"
	
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	grpcopentelemetry "github.com/grpc-ecosystem/go-grpc-middleware/v2/providers/opentelemetry"
)

func Example() {
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			tags.UnaryServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			tracing.UnaryServerInterceptor(grpcopentelemetry.InterceptorTracer()),
		),
		grpc.ChainStreamInterceptor(
			tags.StreamServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			tracing.StreamServerInterceptor(grpcopentelemetry.InterceptorTracer()),
		),
	)
	
	_, _ = grpc.Dial("",
		grpc.WithUnaryInterceptor(tracing.UnaryClientInterceptor(grpcopentelemetry.InterceptorTracer())),
		grpc.WithStreamInterceptor(tracing.StreamClientInterceptor(grpcopentelemetry.InterceptorTracer())),
	)
}

func TestExamplesBuildable(t *testing.T) {
	Example()
}
