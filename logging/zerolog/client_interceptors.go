package zerolog

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog/ctxzr"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"path"
	"time"
)

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
func UnaryClientInterceptor(logger *zerolog.Logger, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		fields := newClientLoggerFields(ctx, method)
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		logFinalClientLine(o, &ctxzr.CtxLogger{Logger: logger, Fields: fields}, startTime, err, "finished client unary call")
		return err
	}
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
func StreamClientInterceptor(logger *zerolog.Logger, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		fields := newClientLoggerFields(ctx, method)
		startTime := time.Now()
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		logFinalClientLine(o, &ctxzr.CtxLogger{Logger: logger, Fields: fields}, startTime, err, "finished client streaming call")
		return clientStream, err
	}
}

func logFinalClientLine(o *options, logger *ctxzr.CtxLogger, startTime time.Time, err error, msg string) {
	code  := o.codeFunc(err)
	level := o.levelFunc(code)

	clientLogger := logger.Logger.WithLevel(level).Err(err)
	args := make(map[string]interface{})
	args["grpc.code"] = code.String()

	for k, v := range logger.Fields {
		args[k] = v
	}
	args["msg"] = msg
	clientLogger.Fields(args)
	// Add Duration to Fields and Send
	o.durationFunc(clientLogger.Fields(args), time.Since(startTime)).Send()
}

func newClientLoggerFields(ctx context.Context, fullMethodString string) map[string]interface{} {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return map[string]interface{}{
		"system":       "grpc",
		"span.kind":    "client",
		"grpc.service": service,
		"grpc.method":  method,
	}
}
