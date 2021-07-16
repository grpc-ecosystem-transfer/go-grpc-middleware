// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package validator

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
)

func TestValidateWrapper(t *testing.T) {
	assert.NoError(t, validate(testpb.GoodPing, false))
	assert.Error(t, validate(testpb.BadPing, false))
	assert.NoError(t, validate(testpb.GoodPing, true))
	assert.Error(t, validate(testpb.BadPing, true))

	assert.NoError(t, validate(testpb.GoodPingError, false))
	assert.Error(t, validate(testpb.BadPingError, false))
	assert.NoError(t, validate(testpb.GoodPingError, true))
	assert.Error(t, validate(testpb.BadPingError, true))

	assert.NoError(t, validate(testpb.GoodPingResponse, false))
	assert.NoError(t, validate(testpb.GoodPingResponse, true))
	assert.Error(t, validate(testpb.BadPingResponse, false))
	assert.Error(t, validate(testpb.BadPingResponse, true))
}

func TestValidatorTestSuite(t *testing.T) {
	s := &ValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(StreamServerInterceptor()),
				grpc.UnaryInterceptor(UnaryServerInterceptor()),
			},
		},
	}
	suite.Run(t, s)
	sAll := &ValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(StreamServerInterceptor(true)),
				grpc.UnaryInterceptor(UnaryServerInterceptor(true)),
			},
		},
	}
	suite.Run(t, sAll)

	cs := &ClientValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ClientOpts: []grpc.DialOption{
				grpc.WithUnaryInterceptor(UnaryClientInterceptor()),
			},
		},
	}
	suite.Run(t, cs)
	csAll := &ClientValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ClientOpts: []grpc.DialOption{
				grpc.WithUnaryInterceptor(UnaryClientInterceptor(true)),
			},
		},
	}
	suite.Run(t, csAll)
}

type ValidatorTestSuite struct {
	*testpb.InterceptorTestSuite
}

func (s *ValidatorTestSuite) TestValidPasses_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	assert.NoError(s.T(), err, "no error expected")
}

func (s *ValidatorTestSuite) TestInvalidErrors_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.BadPing)
	assert.Error(s.T(), err, "no error expected")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func (s *ValidatorTestSuite) TestValidPasses_ServerStream() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	require.NoError(s.T(), err, "no error on stream establishment expected")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		assert.NoError(s.T(), err, "no error on messages sent occurred")
	}
}

func (s *ValidatorTestSuite) TestInvalidErrors_ServerStream() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.BadPingList)
	require.NoError(s.T(), err, "no error on stream establishment expected")
	_, err = stream.Recv()
	assert.Error(s.T(), err, "error should be received on first message")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func (s *ValidatorTestSuite) TestInvalidErrors_BidiStream() {
	stream, err := s.Client.PingStream(s.SimpleCtx())
	require.NoError(s.T(), err, "no error on stream establishment expected")

	require.NoError(s.T(), stream.Send(testpb.GoodPingStream))
	_, err = stream.Recv()
	assert.NoError(s.T(), err, "receiving a good ping should return a good pong")
	require.NoError(s.T(), stream.Send(testpb.GoodPingStream))
	_, err = stream.Recv()
	assert.NoError(s.T(), err, "receiving a good ping should return a good pong")

	require.NoError(s.T(), stream.Send(testpb.BadPingStream))
	_, err = stream.Recv()
	assert.Error(s.T(), err, "receiving a bad ping should return a bad pong")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")

	err = stream.CloseSend()
	assert.NoError(s.T(), err, "there should be no error closing the stream on send")
}

type ClientValidatorTestSuite struct {
	*testpb.InterceptorTestSuite
}

func (s *ClientValidatorTestSuite) TestValidPasses_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	assert.NoError(s.T(), err, "no error expected")
}

func (s *ClientValidatorTestSuite) TestInvalidErrors_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.BadPing)
	assert.Error(s.T(), err, "error expected")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}
