// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package validator

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The validateAller interface at protoc-gen-validate main branch.
// See https://github.com/envoyproxy/protoc-gen-validate/pull/468.
type validateAller interface {
	ValidateAll() error
}

// The validate interface starting with protoc-gen-validate v0.6.0.
// See https://github.com/envoyproxy/protoc-gen-validate/pull/455.
type validator interface {
	Validate(all bool) error
}

// The validate interface prior to protoc-gen-validate v0.6.0.
type validatorLegacy interface {
	Validate() error
}

func validate(req interface{}, all bool) error {
	if all {
		switch v := req.(type) {
		case validateAller:
			if err := v.ValidateAll(); err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
		case validator:
			if err := v.Validate(true); err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
		case validatorLegacy:
			// Fallback to legacy validator
			if err := v.Validate(); err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
		}
		return nil
	}
	switch v := req.(type) {
	case validatorLegacy:
		if err := v.Validate(); err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	case validator:
		if err := v.Validate(false); err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return nil
}

// UnaryServerInterceptor returns a new unary server interceptor that validates incoming messages.
//
// Invalid messages will be rejected with `InvalidArgument` before reaching any userspace handlers.
func UnaryServerInterceptor(validateAll ...bool) grpc.UnaryServerInterceptor {
	var all = false
	if len(validateAll) > 0 {
		all = validateAll[0]
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := validate(req, all); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// UnaryClientInterceptor returns a new unary client interceptor that validates outgoing messages.
//
// Invalid messages will be rejected with `InvalidArgument` before sending the request to server.
func UnaryClientInterceptor(validateAll ...bool) grpc.UnaryClientInterceptor {
	var all = false
	if len(validateAll) > 0 && validateAll[0] {
		all = true
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := validate(req, all); err != nil {
			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that validates incoming messages.
//
// The stage at which invalid messages will be rejected with `InvalidArgument` varies based on the
// type of the RPC. For `ServerStream` (1:m) requests, it will happen before reaching any userspace
// handlers. For `ClientStream` (n:1) or `BidiStream` (n:m) RPCs, the messages will be rejected on
// calls to `stream.Recv()`.
func StreamServerInterceptor(validateAll ...bool) grpc.StreamServerInterceptor {
	var all = false
	if len(validateAll) > 0 && validateAll[0] {
		all = true
	}
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &recvWrapper{
			all:          all,
			ServerStream: stream,
		}
		return handler(srv, wrapper)
	}
}

type recvWrapper struct {
	all bool
	grpc.ServerStream
}

func (s *recvWrapper) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if err := validate(m, s.all); err != nil {
		return err
	}
	return nil
}
