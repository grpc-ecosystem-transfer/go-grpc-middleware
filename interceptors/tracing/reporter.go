package tracing

import (
	"context"
	"io"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing/kv"
)

type reporter struct {
	ctx  context.Context
	span Span

	receivedMessageID int
	sentMessageID     int
}

func (o *reporter) PostCall(err error, _ time.Duration) {
	// Finish span.
	if err != nil && err != io.EOF {
		s, _ := status.FromError(err)
		o.span.SetStatus(s.Code(), s.Message())
		o.span.SetAttributes(statusCodeAttr(s.Code()))
	} else {
		o.span.SetAttributes(statusCodeAttr(codes.OK))
	}
	o.span.End()
}

func (o *reporter) PostMsgSend(payload interface{}, err error, d time.Duration) {
	o.sentMessageID++

	addEvent(o.span, RPCMessageTypeSent, o.sentMessageID, payload)
}

func (o *reporter) PostMsgReceive(payload interface{}, err error, d time.Duration) {
	o.receivedMessageID++

	addEvent(o.span, RPCMessageTypeReceived, o.receivedMessageID, payload)
}

func addEvent(span Span, messageType kv.KeyValue, messageID int, payload interface{}) {
	if p, ok := payload.(proto.Message); ok {
		span.AddEvent("message",
			messageType,
			rpcMessageIDKey.Int(messageID),
			rpcMessageUncompressedSizeKey.Int(proto.Size(p)),
		)
	} else {
		span.AddEvent("message",
			messageType,
			rpcMessageIDKey.Int(messageID),
		)
	}
}

// statusCodeAttr returns status code attribute based on given gRPC code
func statusCodeAttr(c codes.Code) kv.KeyValue {
	return grpcStatusCodeKey.Int64(int64(c))
}