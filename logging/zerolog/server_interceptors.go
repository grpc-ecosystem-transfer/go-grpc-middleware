package grpc_zerolog

import (
	"fmt"
	"github.com/rs/zerolog"
	"path"
	"time"

	"context"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	//"github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog/ctxzr"
	"github.com/Ahmet-Kaplan/go-grpc-middleware/logging/zerolog/ctxzr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	// SystemField is used in every log statement made through grpc_zap. Can be overwritten before any initialization code.
	SystemField = "grpc"
	// ServerField is used in every server-side log statement made through grpc_zap.Can be overwritten before initialization.
	ServerField = "server"
)

// UnaryServerInterceptor returns a new unary server interceptors that adds zap.Logger to the context.
func UnaryServerInterceptor(logger *zerolog.Logger, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		newCtx := injectLogger(ctx, logger, info.FullMethod, startTime)

		resp, err := handler(newCtx, req)
		if !o.shouldLog(info.FullMethod, err) {
			return resp, err
		}

		code := o.codeFunc(err)
		logCall(newCtx, o, "finished unary call with code "+code.String(), code, startTime, err)

		return resp, err
	}
}

func StreamServerInterceptor(logger *zerolog.Logger, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()
		newCtx := injectLogger(stream.Context(), logger, info.FullMethod, startTime)

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx

		err := handler(srv, wrapped)
		if !o.shouldLog(info.FullMethod, err) {
			return err
		}

		code := o.codeFunc(err)
		logCall(newCtx, o, "finished streaming call with code "+code.String(), code, startTime, err)

		return err
	}
}

func injectLogger(ctx context.Context, logger *zerolog.Logger, fullMethodString string, start time.Time) context.Context {
	f := ctxzr.TagsToFields(ctx)
	f = append(f, "grpc.start_time", start.Format(time.RFC3339))
	if d, ok := ctx.Deadline(); ok {
		f = append(f, "grpc.request.deadline", d.Format(time.RFC3339))
	}
	f = append(f, serverCallFields(fullMethodString)...)
	var injectLog = ctxzr.CtxLogger{Logger: logger, Fields: f}

	return ctxzr.ToContext(ctx, &injectLog)
}

func serverCallFields(fullMethodString string) []interface{} {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return []interface{}{
		"system", SystemField,
		"span.kind", ServerField,
		"grpc.service", service,
		"grpc.method", method,
	}
}

func logCall(ctx context.Context, options *options, msg string, code codes.Code, startTime time.Time, err error) {

	extractedLogger := ctxzr.Extract(ctx)

	var level = options.levelFunc(code)
	args := []interface{}{"msg", msg, "error", err, "grpc.code", code.String()}
	args = append(args, options.durationFunc(time.Since(startTime))...)
	extractedLogger.Logger.WithLevel(level).Msg(fmt.Sprint(args...))
}
