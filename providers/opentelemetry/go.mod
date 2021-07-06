module github.com/grpc-ecosystem/go-grpc-middleware/providers/opentelemetry/v2

go 1.14

replace github.com/grpc-ecosystem/go-grpc-middleware/v2 => ../..

require (
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-rc.2
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/trace v0.20.0
	google.golang.org/grpc v1.37.0
)
