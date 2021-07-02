// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_middleware

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedContextSettingServerStream allows setting the wrapped context of a grpc.ServerStream.
type WrappedContextSettingServerStream interface {
	grpc.ServerStream
	SetContext(context.Context)
}

// WrappedServerStream is a thin wrapper around grpc.ServerStream that allows modifying context.
type WrappedServerStream struct {
	grpc.ServerStream
	// WrappedContext is the wrapper's own Context. You can assign it.
	WrappedContext context.Context
}

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *WrappedServerStream) Context() context.Context {
	return w.WrappedContext
}

// SetContext sets the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *WrappedServerStream) SetContext(ctx context.Context) {
	w.WrappedContext = ctx
}

// WrapServerStream returns a ServerStream that has the ability to overwrite context.
func WrapServerStream(stream grpc.ServerStream) WrappedContextSettingServerStream {
	if existing, ok := stream.(WrappedContextSettingServerStream); ok {
		return existing
	}
	return &WrappedServerStream{ServerStream: stream, WrappedContext: stream.Context()}
}
