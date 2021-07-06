package tracing

import (
	"context"
	
	"google.golang.org/grpc/codes"
	
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing/kv"
)

const (
	// Type of message transmitted or received.
	rpcMessageTypeKey = kv.Key("message.type")
	
	// Identifier of message transmitted or received.
	rpcMessageIDKey = kv.Key("message.id")
	
	// The uncompressed size of the message transmitted or received in
	// bytes.
	rpcMessageUncompressedSizeKey = kv.Key("message.uncompressed_size")
	
	// grpcStatusCodeKey is convention for numeric status code of a gRPC request.
	grpcStatusCodeKey = kv.Key("rpc.grpc.status_code")
)

var (
	RPCMessageTypeSent     = rpcMessageTypeKey.String("SENT")
	RPCMessageTypeReceived = rpcMessageTypeKey.String("RECEIVED")
)

type Tracer interface {
	Start(ctx context.Context, spanName string, kind SpanKind) (context.Context, Span)
}

type Span interface {
	// End completes the span. No updates are allowed to span after it
	// ends. The only exception is setting status of the span.
	End()
	
	// SetStatus sets the status of the span in the form of a code
	// and a message. SetStatus overrides the value of previous
	// calls to SetStatus on the Span.
	//
	// The default span status is OK, so it is not necessary to
	// explicitly set an OK status on successful Spans unless it
	// is to add an OK message or to override a previous status on the Span.
	SetStatus(code codes.Code, msg string)
	
	// AddEvent adds an event to the span.
	// Middleware will call it while receiving or sending messages.
	AddEvent(name string, attrs ...kv.KeyValue)
	
	// SetAttributes sets kv as attributes of the Span. If a key from kv
	// already exists for an attribute of the Span it should be overwritten with
	// the value contained in kv.
	SetAttributes(attrs ...kv.KeyValue)
}
