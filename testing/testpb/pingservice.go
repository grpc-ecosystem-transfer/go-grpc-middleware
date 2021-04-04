// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

/*
Package `grpc_testing` provides helper functions for testing validators in this package.
*/

package testpb

import (
	"context"
	"io"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// ListResponseCount is the expected number of responses to PingList
	ListResponseCount   = 100
	TestServiceFullName = "testing.testpb.v1.TestService"
)

// Interface implementation assert.
var _ TestServiceServer = &TestPingService{}

type TestPingService struct {
	UnimplementedTestServiceServer

	T *testing.T
}

func (s *TestPingService) PingEmpty(_ context.Context, _ *PingEmptyRequest) (*PingEmptyResponse, error) {
	return &PingEmptyResponse{}, nil
}

func (s *TestPingService) Ping(_ context.Context, ping *PingRequest) (*PingResponse, error) {
	// Send user trailers and headers.
	return &PingResponse{Value: ping.Value, Counter: 0}, nil
}

func (s *TestPingService) PingError(_ context.Context, ping *PingErrorRequest) (*PingErrorResponse, error) {
	code := codes.Code(ping.ErrorCodeReturned)
	return nil, status.Errorf(code, "Userspace error.")
}

func (s *TestPingService) PingList(ping *PingListRequest, stream TestService_PingListServer) error {
	if ping.ErrorCodeReturned != 0 {
		return status.Errorf(codes.Code(ping.ErrorCodeReturned), "foobar")
	}

	// Send user trailers and headers.
	for i := 0; i < ListResponseCount; i++ {
		if err := stream.Send(&PingListResponse{Value: ping.Value, Counter: int32(i)}); err != nil {
			return err
		}
	}
	return nil
}

func (s *TestPingService) PingStream(stream TestService_PingStreamServer) error {
	count := 0
	for {
		ping, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&PingStreamResponse{Value: ping.Value, Counter: int32(count)}); err != nil {
			return err
		}

		count += 1
	}
	return nil
}
