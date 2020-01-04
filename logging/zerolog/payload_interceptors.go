package grpc_zerolog

import (
	"bytes"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog/ctxzr"

	"context"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

var (
	// JsonPbMarshaller is the marshaller used for serializing protobuf messages.
	JsonPbMarshaller = &jsonpb.Marshaler{}
)

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests.
//
// This *only* works when placed *after* the `grpc_zerolog.UnaryServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadUnaryServerInterceptor(logger *zerolog.Logger, decider grpc_logging.ServerPayloadLoggingDecider) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !decider(ctx, info.FullMethod, info.Server) {
			return handler(ctx, req)
		}
		// Use the provided log.Logger for logging but use the fields from context.
		resLogger := ctxzr.CtxLogger{Logger: logger, Fields: ctxzr.MergeFields(serverCallFields(info.FullMethod), ctxzr.TagsToFields(ctx))}

		logProtoMessageAsJson(&resLogger, req, "grpc.request.content", "server request payload logged as grpc.request.content field")
		resp, err := handler(ctx, req)
		if err == nil {
			logProtoMessageAsJson(&resLogger, resp, "grpc.response.content", "server response payload logged as grpc.response.content field")
		}
		return resp, err
	}
}

// PayloadStreamServerInterceptor returns a new server server interceptors that logs the payloads of requests.
//
// This *only* works when placed *after* the `grpc_zerolog.StreamServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadStreamServerInterceptor(logger *zerolog.Logger, decider grpc_logging.ServerPayloadLoggingDecider) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !decider(stream.Context(), info.FullMethod, srv) {
			return handler(srv, stream)
		}
		logEntry := ctxzr.CtxLogger{Logger: logger, Fields: ctxzr.MergeFields(serverCallFields(info.FullMethod), ctxzr.TagsToFields(stream.Context()))}

		newStream := &loggingServerStream{ServerStream: stream, logger: &logEntry}
		return handler(srv, newStream)
	}
}

// PayloadUnaryClientInterceptor returns a new unary client interceptor that logs the paylods of requests and responses.
func PayloadUnaryClientInterceptor(logger *zerolog.Logger, decider grpc_logging.ClientPayloadLoggingDecider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !decider(ctx, method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		logEntry := ctxzr.CtxLogger{Logger: logger, Fields: newClientLoggerFields(ctx, method)}
		logProtoMessageAsJson(&logEntry, req, "grpc.request.content", "client request payload logged as grpc.request.content")
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			logProtoMessageAsJson(&logEntry, reply, "grpc.response.content", "client response payload logged as grpc.response.content")
		}
		return err
	}
}

// PayloadStreamClientInterceptor returns a new streaming client interceptor that logs the paylods of requests and responses.
func PayloadStreamClientInterceptor(logger *zerolog.Logger, decider grpc_logging.ClientPayloadLoggingDecider) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if !decider(ctx, method) {
			return streamer(ctx, desc, cc, method, opts...)
		}
		logEntry := ctxzr.CtxLogger{Logger: logger, Fields: newClientLoggerFields(ctx, method)}
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		newStream := &loggingClientStream{ClientStream: clientStream, logger: &logEntry}
		return newStream, err
	}
}

type loggingClientStream struct {
	grpc.ClientStream
	logger *ctxzr.CtxLogger
}

func (l *loggingClientStream) SendMsg(m interface{}) error {
	err := l.ClientStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, m, "grpc.request.content", "server request payload logged as grpc.request.content field")
	}
	return err
}

func (l *loggingClientStream) RecvMsg(m interface{}) error {
	err := l.ClientStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, m, "grpc.response.content", "server response payload logged as grpc.response.content field")
	}
	return err
}

type loggingServerStream struct {
	grpc.ServerStream
	logger *ctxzr.CtxLogger
}

func (l *loggingServerStream) SendMsg(m interface{}) error {
	err := l.ServerStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, m, "grpc.response.content", "server response payload logged as grpc.response.content field")
	}
	return err
}

func (l *loggingServerStream) RecvMsg(m interface{}) error {
	err := l.ServerStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, m, "grpc.request.content", "server request payload logged as grpc.request.content field")
	}
	return err
}

func logProtoMessageAsJson(logger *ctxzr.CtxLogger, pbMsg interface{}, key string, msg string) {
	if p, ok := pbMsg.(proto.Message); ok {
		payload, err := (&jsonpbObjectMarshaler{pb: p}).MarshalJSON()
		if err != nil {
			logger.Logger.Err(err).Fields(logger.Fields).RawJSON(key, payload).Msg(msg)
		} else {
			logger.Logger.Info().Fields(logger.Fields).RawJSON(key, payload).Msg(msg)
		}
	}
}

type jsonpbObjectMarshaler struct {
	pb proto.Message
}

func (j *jsonpbObjectMarshaler) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := JsonPbMarshaller.Marshal(b, j.pb); err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}
	return b.Bytes(), nil
}
