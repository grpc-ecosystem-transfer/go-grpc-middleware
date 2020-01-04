package ctxzr_test

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog/ctxzr"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
)

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func ExampleExtract_unary() {
	_ = func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
		// Add fields the ctxtags of the request which will be added to all extracted loggers.
		grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)

		// Extract a single request-scoped zap.Logger and log messages.
		l := ctxzr.Extract(ctx)
		l.Fields["msg1"] = "some ping"
		l.Fields["msg1"] = "another ping"
		l.Logger.Info().Fields(l.Fields).Send()
		return &pb_testproto.PingResponse{Value: ping.Value}, nil
	}
}
