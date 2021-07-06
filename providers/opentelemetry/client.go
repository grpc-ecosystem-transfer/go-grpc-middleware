package opentelemetry

import (
	"context"
	"net"
	"strings"
	
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
	
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

func newClientSpanFromContext(ctx context.Context, fullMethod string, tracer trace.Tracer, propagators propagation.TextMapPropagator) (context.Context, trace.Span) {
	name, attrs := spanInfo(ctx, fullMethod)
	
	ctx, span := tracer.Start(
		ctx,
		name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
	
	requestMetadata, _ := metadata.FromOutgoingContext(ctx)
	metadataCopy := requestMetadata.Copy()
	otelgrpc.Inject(ctx, &metadataCopy, otelgrpc.WithPropagators(propagators))
	ctx = metadata.NewOutgoingContext(ctx, metadataCopy)
	
	return ctx, span
}

func spanInfo(ctx context.Context, fullMethod string) (string, []attribute.KeyValue) {
	t := tags.Extract(ctx)
	
	attrs := []attribute.KeyValue{semconv.RPCSystemGRPC}
	name, mAttrs := parseFullMethod(fullMethod)
	attrs = append(attrs, mAttrs...)
	attrs = append(attrs, peerAttr(t.Values()["peer.address"])...)
	return name, attrs
}

// peerAttr returns attributes about the peer address.
func peerAttr(addr string) []attribute.KeyValue {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return []attribute.KeyValue(nil)
	}
	
	if host == "" {
		host = "127.0.0.1"
	}
	
	return []attribute.KeyValue{
		semconv.NetPeerIPKey.String(host),
		semconv.NetPeerPortKey.String(port),
	}
}

// parseFullMethod returns a span name following the OpenTelemetry semantic
// conventions as well as all applicable span kv.KeyValue attributes based
// on a gRPC's FullMethod.
func parseFullMethod(fullMethod string) (string, []attribute.KeyValue) {
	// The span name MUST be the full RPC method name formatted as
	// `$package.$service/$method`.
	name := strings.TrimLeft(fullMethod, "/")
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		// Invalid format, does not follow `$package.$service/$method`.
		return name, []attribute.KeyValue(nil)
	}
	
	var attrs []attribute.KeyValue
	if service := parts[0]; service != "" {
		attrs = append(attrs, semconv.RPCServiceKey.String(service))
	}
	if method := parts[1]; method != "" {
		attrs = append(attrs, semconv.RPCMethodKey.String(method))
	}
	return name, attrs
}
