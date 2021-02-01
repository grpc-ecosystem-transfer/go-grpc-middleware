package metrics

import (
	"context"
	"io"
	"log"
	"net"
	"testing"
	"time"

	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2/testproto/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	DefaultClientMetrics *ClientMetrics = NewClientMetrics(prometheus.DefaultRegisterer)
)

func TestClientInterceptorSuite(t *testing.T) {
	suite.Run(t, &ClientInterceptorTestSuite{})
}

type ClientInterceptorTestSuite struct {
	suite.Suite

	serverListener net.Listener
	server         *grpc.Server
	clientConn     *grpc.ClientConn
	testClient     pb_testproto.TestServiceClient
	ctx            context.Context
	cancel         context.CancelFunc
}

func (s *ClientInterceptorTestSuite) SetupSuite() {
	var err error

	// Make all RPC calls last at most 2 sec, meaning all async issues or deadlock will not kill tests.
	s.ctx, s.cancel = context.WithTimeout(context.TODO(), 2*time.Second)
	_ = DefaultClientMetrics.EnableClientHandlingTimeHistogram()

	s.serverListener, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err, "must be able to allocate a port for serverListener")

	// This is the point where we hook up the interceptor
	s.server = grpc.NewServer()
	pb_testproto.RegisterTestServiceServer(s.server, &testService{t: s.T()})

	go func() {
		if err := s.server.Serve(s.serverListener); err != nil {
			log.Fatalf("%v", err)
		}
	}()

	s.clientConn, err = grpc.DialContext(
		s.ctx,
		s.serverListener.Addr().String(),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(UnaryClientInterceptor(prometheus.DefaultRegisterer)),
		grpc.WithStreamInterceptor(StreamClientInterceptor(prometheus.DefaultRegisterer)))
	require.NoError(s.T(), err, "must not error on client Dial")
	s.testClient = pb_testproto.NewTestServiceClient(s.clientConn)
}

func (s *ClientInterceptorTestSuite) SetupTest() {

	// Make sure every test starts with same fresh, intialized metric state.
	DefaultClientMetrics.clientStartedCounter.Reset()
	DefaultClientMetrics.clientHandledCounter.Reset()
	DefaultClientMetrics.clientHandledHistogram.Reset()
	DefaultClientMetrics.clientStreamMsgReceived.Reset()
	DefaultClientMetrics.clientStreamMsgSent.Reset()
}

func (s *ClientInterceptorTestSuite) TearDownSuite() {
	if s.serverListener != nil {
		s.server.Stop()
		s.T().Logf("stopped grpc.Server at: %v", s.serverListener.Addr().String())
		s.serverListener.Close()

	}
	if s.clientConn != nil {
		s.clientConn.Close()
	}
}

func (s *ClientInterceptorTestSuite) TearDownTest() {
	s.cancel()
}

func (s *ClientInterceptorTestSuite) TestUnaryIncrementsMetrics() {
	_, err := s.testClient.PingEmpty(s.ctx, &pb_testproto.PingEmptyRequest{}) // should return with code=OK
	require.NoError(s.T(), err)
	requireValue(s.T(), 1, DefaultClientMetrics.clientStartedCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingEmpty"))
	requireValue(s.T(), 1, DefaultClientMetrics.clientHandledCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingEmpty", "OK"))
	requireValueHistCount(s.T(), 1, DefaultClientMetrics.clientHandledHistogram.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingEmpty"))

	_, err = s.testClient.PingError(s.ctx, &pb_testproto.PingErrorRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)}) // should return with code=FailedPrecondition
	require.Error(s.T(), err)
	requireValue(s.T(), 1, DefaultClientMetrics.clientStartedCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingError"))
	requireValue(s.T(), 1, DefaultClientMetrics.clientHandledCounter.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingError", "FailedPrecondition"))
	requireValueHistCount(s.T(), 1, DefaultClientMetrics.clientHandledHistogram.WithLabelValues("unary", "mwitkow.testproto.TestService", "PingError"))
}

func (s *ClientInterceptorTestSuite) TestStartedStreamingIncrementsStarted() {
	_, err := s.testClient.PingList(s.ctx, &pb_testproto.PingListRequest{})
	require.NoError(s.T(), err)
	requireValue(s.T(), 1, DefaultClientMetrics.clientStartedCounter.WithLabelValues("server_stream", "mwitkow.testproto.TestService", "PingList"))

	_, err = s.testClient.PingList(s.ctx, &pb_testproto.PingListRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)}) // should return with code=FailedPrecondition
	require.NoError(s.T(), err, "PingList must not fail immediately")
	requireValue(s.T(), 2, DefaultClientMetrics.clientStartedCounter.WithLabelValues("server_stream", "mwitkow.testproto.TestService", "PingList"))
}

func (s *ClientInterceptorTestSuite) TestStreamingIncrementsMetrics() {
	ss, _ := s.testClient.PingList(s.ctx, &pb_testproto.PingListRequest{}) // should return with code=OK
	// Do a read, just for kicks.
	count := 0
	for {
		_, err := ss.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading pingList shouldn't fail")
		count++
	}
	require.EqualValues(s.T(), countListResponses, count, "Number of received msg on the wire must match")

	requireValue(s.T(), 1, DefaultClientMetrics.clientStartedCounter.WithLabelValues("server_stream", "mwitkow.testproto.TestService", "PingList"))
	requireValue(s.T(), 1, DefaultClientMetrics.clientHandledCounter.WithLabelValues("server_stream", "mwitkow.testproto.TestService", "PingList", "OK"))
	requireValue(s.T(), countListResponses, DefaultClientMetrics.clientStreamMsgReceived.WithLabelValues("server_stream", "mwitkow.testproto.TestService", "PingList"))
	requireValue(s.T(), 1, DefaultClientMetrics.clientStreamMsgSent.WithLabelValues("server_stream", "mwitkow.testproto.TestService", "PingList"))
	requireValueHistCount(s.T(), 1, DefaultClientMetrics.clientHandledHistogram.WithLabelValues("server_stream", "mwitkow.testproto.TestService", "PingList"))

	ss, err := s.testClient.PingList(s.ctx, &pb_testproto.PingListRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)}) // should return with code=FailedPrecondition
	require.NoError(s.T(), err, "PingList must not fail immediately")

	// Do a read, just to progate errors.
	_, err = ss.Recv()
	st, _ := status.FromError(err)
	require.Equal(s.T(), codes.FailedPrecondition, st.Code(), "Recv must return FailedPrecondition, otherwise the test is wrong")

	requireValue(s.T(), 2, DefaultClientMetrics.clientStartedCounter.WithLabelValues("server_stream", "mwitkow.testproto.TestService", "PingList"))
	requireValue(s.T(), 1, DefaultClientMetrics.clientHandledCounter.WithLabelValues("server_stream", "mwitkow.testproto.TestService", "PingList", "FailedPrecondition"))
	requireValueHistCount(s.T(), 2, DefaultClientMetrics.clientHandledHistogram.WithLabelValues("server_stream", "mwitkow.testproto.TestService", "PingList"))
}
