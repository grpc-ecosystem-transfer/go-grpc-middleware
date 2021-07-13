// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logging

import (
	"bytes"
	"context"
	"fmt"
	"time"

	//lint:ignore SA1019 we use this deprecated package to convert pb v1 messages to pb v2
	//nolint:staticcheck // SA1019
	protov1 "github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

type serverPayloadReporter struct {
	ctx    context.Context
	logger Logger
}

func (c *serverPayloadReporter) PostCall(error, time.Duration) {}

func (c *serverPayloadReporter) PostMsgSend(req interface{}, err error, duration time.Duration) {
	if err != nil {
		return
	}

	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	p, ok := message(req)
	if !ok {
		logger.With("req.type", fmt.Sprintf("%T", req)).Log(ERROR, "req is not a protocol buffers message; programmatic error?")

		return
	}
	// For server send message is the response.
	logProtoMessageAsJson(logger.With("grpc.send.duration", duration.String()), p, "grpc.response.content", "response payload logged as grpc.response.content field")
}

func (c *serverPayloadReporter) PostMsgReceive(reply interface{}, err error, duration time.Duration) {
	if err != nil {
		return
	}

	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)

	p, ok := message(reply)
	if !ok {
		logger.With("reply.type", fmt.Sprintf("%T", reply)).Log(ERROR, "reply is not a protocol buffers message; programmatic error?")
		return
	}
	// For server recv message is the request.
	logProtoMessageAsJson(logger.With("grpc.recv.duration", duration.String()), p, "grpc.request.content", "request payload logged as grpc.request.content field")
}

type clientPayloadReporter struct {
	ctx    context.Context
	logger Logger
}

func (c *clientPayloadReporter) PostCall(error, time.Duration) {}

func (c *clientPayloadReporter) PostMsgSend(req interface{}, err error, duration time.Duration) {
	if err != nil {
		return
	}

	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	p, ok := message(req)
	if !ok {
		logger.With("req.type", fmt.Sprintf("%T", req)).Log(ERROR, "req is not a protocol buffers message; programmatic error?")
		return
	}
	logProtoMessageAsJson(logger.With("grpc.send.duration", duration.String()), p, "grpc.request.content", "request payload logged as grpc.request.content field")
}

func (c *clientPayloadReporter) PostMsgReceive(reply interface{}, err error, duration time.Duration) {
	if err != nil {
		return
	}

	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	p, ok := message(reply)
	if !ok {
		logger.With("reply.type", fmt.Sprintf("%T", reply)).Log(ERROR, "reply is not a protocol buffers message; programmatic error?")
		return
	}
	logProtoMessageAsJson(logger.With("grpc.recv.duration", duration.String()), p, "grpc.response.content", "response payload logged as grpc.response.content field")
}

type payloadReportable struct {
	clientDecider   ClientPayloadLoggingDecider
	serverDecider   ServerPayloadLoggingDecider
	logger          Logger
	timestampFormat string
}

func (r *payloadReportable) ServerReporter(ctx context.Context, req interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	if !r.serverDecider(ctx, interceptors.FullMethod(service, method), req) {
		return interceptors.NoopReporter{}, ctx
	}
	fields := commonFields(KindServerFieldValue, typ, service, method)
	fields = append(fields, "grpc.start_time", time.Now().Format(r.timestampFormat))
	if d, ok := ctx.Deadline(); ok {
		fields = append(fields, "grpc.request.deadline", d.Format(r.timestampFormat))
	}
	return &serverPayloadReporter{
		ctx:    ctx,
		logger: r.logger.With(fields...),
	}, ctx
}
func (r *payloadReportable) ClientReporter(ctx context.Context, _ interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	if !r.clientDecider(ctx, interceptors.FullMethod(service, method)) {
		return interceptors.NoopReporter{}, ctx
	}
	fields := commonFields(KindClientFieldValue, typ, service, method)
	fields = append(fields, "grpc.start_time", time.Now().Format(r.timestampFormat))
	if d, ok := ctx.Deadline(); ok {
		fields = append(fields, "grpc.request.deadline", d.Format(r.timestampFormat))
	}
	return &clientPayloadReporter{
		ctx:    ctx,
		logger: r.logger.With(fields...),
	}, ctx
}

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests on INFO level.
// Logger tags will be used from tags context.
func PayloadUnaryServerInterceptor(logger Logger, decider ServerPayloadLoggingDecider, timestampFormat string) grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(&payloadReportable{logger: logger, serverDecider: decider, timestampFormat: timestampFormat})
}

// PayloadStreamServerInterceptor returns a new server server interceptors that logs the payloads of requests on INFO level.
// Logger tags will be used from tags context.
func PayloadStreamServerInterceptor(logger Logger, decider ServerPayloadLoggingDecider, timestampFormat string) grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(&payloadReportable{logger: logger, serverDecider: decider, timestampFormat: timestampFormat})
}

// PayloadUnaryClientInterceptor returns a new unary client interceptor that logs the payloads of requests and responses on INFO level.
// Logger tags will be used from tags context.
func PayloadUnaryClientInterceptor(logger Logger, decider ClientPayloadLoggingDecider, timestampFormat string) grpc.UnaryClientInterceptor {
	return interceptors.UnaryClientInterceptor(&payloadReportable{logger: logger, clientDecider: decider, timestampFormat: timestampFormat})
}

// PayloadStreamClientInterceptor returns a new streaming client interceptor that logs the paylods of requests and responses on INFO level.
// Logger tags will be used from tags context.
func PayloadStreamClientInterceptor(logger Logger, decider ClientPayloadLoggingDecider, timestampFormat string) grpc.StreamClientInterceptor {
	return interceptors.StreamClientInterceptor(&payloadReportable{logger: logger, clientDecider: decider, timestampFormat: timestampFormat})
}

func message(msg interface{}) (proto.Message, bool) {
	pbMsg, ok := msg.(proto.Message)
	if !ok {
		pbV1, v1ok := msg.(protov1.Message)
		if !v1ok {
			return nil, false
		}
		pbMsg = protov1.MessageV2(pbV1)
	}
	return pbMsg, true
}

func logProtoMessageAsJson(logger Logger, pbMsg proto.Message, key string, msg string) {
	payload, err := protojson.Marshal(pbMsg)
	if err != nil {
		logger = logger.With(key, err.Error())
	} else {
		// Trim spaces for deterministic output.
		// See: https://github.com/golang/protobuf/issues/1269
		logger = logger.With(key, string(bytes.Replace(payload, []byte{' '}, []byte{}, -1)))
	}
	logger.Log(INFO, msg)
}
