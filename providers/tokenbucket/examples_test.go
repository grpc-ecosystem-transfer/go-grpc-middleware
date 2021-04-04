package tokenbucket

import (
	"github.com/juju/ratelimit"
	"google.golang.org/grpc"

	grpc_ratelimit "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
)

// Hard-coded for simplicity sake, but make this configurable in your application.
const (
	// Add 5 token per seconds.
	rate = 5
	// Capacity of bucket. allow only 40 requests.
	tokenCapacity = 40
)

// Simple example of server initialization code.
func Example() {
	limiter := TokenBucketInterceptor{}
	limiter.tokenBucket = ratelimit.NewBucket(rate, int64(tokenCapacity))

	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_ratelimit.UnaryServerInterceptor(&limiter),
		),
		grpc.ChainStreamInterceptor(
			grpc_ratelimit.StreamServerInterceptor(&limiter),
		),
	)
}
