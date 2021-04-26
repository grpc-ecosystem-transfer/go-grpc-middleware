// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package tokenbucket

// Implement Limiter interface.

import (
	"context"
	"fmt"

	"github.com/juju/ratelimit"
)

// TokenBucketInterceptor implement token bucket algorithm.
type TokenBucketInterceptor struct {
	tokenBucket *ratelimit.Bucket
}

// Limit implements Limiter interface.
func (r *TokenBucketInterceptor) Limit(_ context.Context) error {
	// Take one token per request. This call doesn't block.
	tokenRes := r.tokenBucket.TakeAvailable(1)

	// When rate limit reached, return specific error for the clients.
	if tokenRes == 0 {
		return fmt.Errorf("APP-XXX: reached Rate-Limiting %d", r.tokenBucket.Available())
	}

	// Rate limit isn't reached.
	return nil
}
