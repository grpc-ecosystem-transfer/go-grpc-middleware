package logging_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
)

type loggingPayloadSuite struct {
	*baseLoggingSuite
}

func TestPayloadSuite(t *testing.T) {
	alwaysLoggingDeciderServer := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }
	alwaysLoggingDeciderClient := func(ctx context.Context, fullMethodName string) bool { return true }

	s := &loggingPayloadSuite{
		baseLoggingSuite: &baseLoggingSuite{
			logger: newMockLogger(),
			InterceptorTestSuite: &testpb.InterceptorTestSuite{
				TestService: &testpb.TestPingService{T: t},
			},
		},
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(logging.PayloadUnaryClientInterceptor(s.logger, alwaysLoggingDeciderClient)),
		grpc.WithStreamInterceptor(logging.PayloadStreamClientInterceptor(s.logger, alwaysLoggingDeciderClient)),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.StreamInterceptor(logging.PayloadStreamServerInterceptor(s.logger, alwaysLoggingDeciderServer)),
		grpc.UnaryInterceptor(logging.PayloadUnaryServerInterceptor(s.logger, alwaysLoggingDeciderServer)),
	}
	suite.Run(t, s)
}

func (s *loggingPayloadSuite) TestPing_LogsBothRequestAndResponse() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")

	lines := s.logger.o.Lines()
	require.Len(s.T(), lines, 4)
	s.assertPayloadLogLinesForMessage(lines, "Ping", interceptors.Unary)
}

func (s *loggingPayloadSuite) assertPayloadLogLinesForMessage(lines LogLines, method string, typ interceptors.GRPCType) {
	// Order matter for assertion, we don't rely on it though, so sort.
	sort.Sort(lines)

	repetitions := len(lines) / 4
	curr := 0
	for i := curr; i < repetitions; i++ {
		clientRequestLogLine := lines[i]
		assert.Equal(s.T(), logging.INFO, clientRequestLogLine.lvl)
		assert.Equal(s.T(), "request payload logged as grpc.request.content field", clientRequestLogLine.msg)
		clientRequestFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientRequestLogLine.fields, method, typ)
		clientRequestFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
			AssertFieldNotEmpty(s.T(), "grpc.send.duration").
			AssertField(s.T(), "grpc.request.content", `{"value":"something","sleepTimeMs":9999}`).
			AssertFieldNotEmpty(s.T(), "grpc.request.deadline").AssertNoMoreTags(s.T())
	}
	curr += repetitions
	for i := curr; i < curr+repetitions; i++ {
		clientResponseLogLine := lines[i]
		assert.Equal(s.T(), logging.INFO, clientResponseLogLine.lvl)
		assert.Equal(s.T(), "response payload logged as grpc.response.content field", clientResponseLogLine.msg)
		clientResponseFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientResponseLogLine.fields, method, typ)
		clientResponseFields = clientResponseFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
			AssertFieldNotEmpty(s.T(), "grpc.recv.duration").
			AssertFieldNotEmpty(s.T(), "grpc.request.deadline")
		if i-curr == 0 {
			clientResponseFields = clientResponseFields.AssertField(s.T(), "grpc.response.content", `{"value":"something"}`)
		} else {
			clientResponseFields = clientResponseFields.AssertField(s.T(), "grpc.response.content", fmt.Sprintf(`{"value":"something","counter":%v}`, i-curr))
		}
		clientResponseFields.AssertNoMoreTags(s.T())
	}
	curr += repetitions
	for i := curr; i < curr+repetitions; i++ {
		serverRequestLogLine := lines[i]
		assert.Equal(s.T(), logging.INFO, serverRequestLogLine.lvl)
		assert.Equal(s.T(), "request payload logged as grpc.request.content field", serverRequestLogLine.msg)
		serverRequestFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverRequestLogLine.fields, method, typ)
		serverRequestFields.AssertFieldNotEmpty(s.T(), "peer.address").
			AssertFieldNotEmpty(s.T(), "grpc.start_time").
			AssertFieldNotEmpty(s.T(), "grpc.recv.duration").
			AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
			AssertField(s.T(), "grpc.request.content", `{"value":"something","sleepTimeMs":9999}`).AssertNoMoreTags(s.T())
	}
	curr += repetitions
	for i := curr; i < curr+repetitions; i++ {
		serverResponseLogLine := lines[i]
		assert.Equal(s.T(), logging.INFO, serverResponseLogLine.lvl)
		assert.Equal(s.T(), "response payload logged as grpc.response.content field", serverResponseLogLine.msg)
		serverResponseFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverResponseLogLine.fields, method, typ)
		serverResponseFields = serverResponseFields.AssertFieldNotEmpty(s.T(), "peer.address").
			AssertFieldNotEmpty(s.T(), "grpc.start_time").
			AssertFieldNotEmpty(s.T(), "grpc.send.duration").
			AssertFieldNotEmpty(s.T(), "grpc.request.deadline")
		if i-curr == 0 {
			serverResponseFields = serverResponseFields.AssertField(s.T(), "grpc.response.content", `{"value":"something"}`)
		} else {
			serverResponseFields = serverResponseFields.AssertField(s.T(), "grpc.response.content", fmt.Sprintf(`{"value":"something","counter":%v}`, i-curr))
		}
		serverResponseFields.AssertNoMoreTags(s.T())
	}
}

func (s *loggingPayloadSuite) TestPingError_LogsOnlyRequestsOnError() {
	_, err := s.Client.PingError(s.SimpleCtx(), &testpb.PingErrorRequest{Value: "something", ErrorCodeReturned: uint32(4)})
	require.Error(s.T(), err, "there must be an error on an unsuccessful call")

	lines := s.logger.o.Lines()
	sort.Sort(lines)
	require.Len(s.T(), lines, 2) // Only client & server requests.
	clientRequestLogLine := lines[0]
	assert.Equal(s.T(), logging.INFO, clientRequestLogLine.lvl)
	assert.Equal(s.T(), "request payload logged as grpc.request.content field", clientRequestLogLine.msg)
	clientRequestFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientRequestLogLine.fields, "PingError", interceptors.Unary)

	clientRequestFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.send.duration").
		AssertField(s.T(), "grpc.request.content", `{"value":"something","errorCodeReturned":4}`).
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").AssertNoMoreTags(s.T())
}

func (s *loggingPayloadSuite) TestPingStream_LogsAllRequestsAndResponses() {
	messagesExpected := 20
	stream, err := s.Client.PingStream(s.SimpleCtx())

	require.NoError(s.T(), err, "no error on stream creation")
	for i := 0; i < messagesExpected; i++ {
		require.NoError(s.T(), stream.Send(testpb.GoodPingStream), "sending must succeed")

		pong := &testpb.PingResponse{}
		err := stream.RecvMsg(pong)
		require.NoError(s.T(), err, "no error on receive")
	}
	require.NoError(s.T(), stream.CloseSend(), "no error on send stream")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	require.NoError(s.T(), waitUntil(200*time.Millisecond, ctx.Done(), func() error {
		got := len(s.logger.o.Lines())
		if got >= 4*messagesExpected {
			return nil
		}
		return errors.Errorf("not enough log lines, waiting; got: %v", got)
	}))
	s.assertPayloadLogLinesForMessage(s.logger.o.Lines(), "PingStream", interceptors.BidiStream)
}

// waitUntil executes f every interval seconds until timeout or no error is returned from f.
func waitUntil(interval time.Duration, stopc <-chan struct{}, f func() error) error {
	tick := time.NewTicker(interval)
	defer tick.Stop()

	var err error
	for {
		if err = f(); err == nil {
			return nil
		}
		select {
		case <-stopc:
			return err
		case <-tick.C:
		}
	}
}
