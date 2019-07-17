// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_validator_test

import (
	"io"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// See test.manual_validator.pb.go for the validator check of SleepTimeMs.
	goodPing = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
	badPing  = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 10001}
)

func TestValidatorTestSuite(t *testing.T) {
	s := &ValidatorTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(grpc_validator.StreamServerInterceptor()),
				grpc.UnaryInterceptor(grpc_validator.UnaryServerInterceptor()),
			},
		},
	}
	suite.Run(t, s)
}

type ValidatorTestSuite struct {
	*grpc_testing.InterceptorTestSuite
}

func (s *ValidatorTestSuite) TestValidPasses_Unary() {
	ctx, cancel := s.SimpleCtx()
	defer cancel()
	_, err := s.Client.Ping(ctx, goodPing)
	assert.NoError(s.T(), err, "no error expected")
}

func (s *ValidatorTestSuite) TestInvalidErrors_Unary() {
	ctx, cancel := s.SimpleCtx()
	defer cancel()
	_, err := s.Client.Ping(ctx, badPing)
	assert.Error(s.T(), err, "no error expected")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func (s *ValidatorTestSuite) TestValidPasses_ServerStream() {
	ctx, cancel := s.SimpleCtx()
	defer cancel()
	stream, err := s.Client.PingList(ctx, goodPing)
	require.NoError(s.T(), err, "no error on stream establishment expected")
	for true {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		assert.NoError(s.T(), err, "no error on messages sent occured")
	}
}

func (s *ValidatorTestSuite) TestInvalidErrors_ServerStream() {
	ctx, cancel := s.SimpleCtx()
	defer cancel()
	stream, err := s.Client.PingList(ctx, badPing)
	require.NoError(s.T(), err, "no error on stream establishment expected")
	_, err = stream.Recv()
	assert.Error(s.T(), err, "error should be received on first message")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func (s *ValidatorTestSuite) TestInvalidErrors_BidiStream() {
	ctx, cancel := s.SimpleCtx()
	defer cancel()
	stream, err := s.Client.PingStream(ctx)
	require.NoError(s.T(), err, "no error on stream establishment expected")

	_ = stream.Send(goodPing)
	_, err = stream.Recv()
	assert.NoError(s.T(), err, "receiving a good ping should return a good pong")
	_ =stream.Send(goodPing)
	_, err = stream.Recv()
	assert.NoError(s.T(), err, "receiving a good ping should return a good pong")

	_ = stream.Send(badPing)
	_, err = stream.Recv()
	assert.Error(s.T(), err, "receiving a good ping should return a good pong")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")

	err = stream.CloseSend()
	assert.NoError(s.T(), err, "there should be no error closing the stream on send")
}
