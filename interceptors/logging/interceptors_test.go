package logging_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
)

type testDisposableFields map[string]string

func (f testDisposableFields) AssertField(t *testing.T, key, value string) testDisposableFields {
	require.Truef(t, len(f) > 0, "expected field %s = %v, but fields ended", key, value)
	assert.Equalf(t, value, f[key], "expected %s for %s", value, key)
	delete(f, key)
	return f
}

func (f testDisposableFields) AssertFieldNotEmpty(t *testing.T, key string) testDisposableFields {
	require.Truef(t, len(f) > 0, "expected field %s and some non-empty value, but fields ended", key)
	assert.Truef(t, f[key] != "", "%s is empty", key)
	delete(f, key)
	return f
}

func (f testDisposableFields) AssertNoMoreTags(t *testing.T) {
	assert.Lenf(t, f, 0, "expected no more fields in testDisposableFields but still got %v", f)
}

type LogLine struct {
	msg    string
	fields testDisposableFields
	lvl    logging.Level
}

type LogLines []LogLine

func (l LogLines) Len() int {
	return len(l)
}

func (l LogLines) Less(i, j int) bool {
	if l[i].fields[logging.ComponentFieldKey] != l[j].fields[logging.ComponentFieldKey] {
		return l[i].fields[logging.ComponentFieldKey] < l[j].fields[logging.ComponentFieldKey]
	}
	if l[i].msg != l[j].msg {
		return l[i].msg < l[j].msg
	}

	// We want to sort by counter which in string, so we need to parse it.
	a := testpb.PingResponse{}
	b := testpb.PingResponse{}
	_ = json.Unmarshal([]byte(l[i].fields["grpc.response.content"]), &a)
	_ = json.Unmarshal([]byte(l[j].fields["grpc.response.content"]), &b)
	if a.Counter != b.Counter {
		return a.Counter < b.Counter
	}

	_ = json.Unmarshal([]byte(l[i].fields["grpc.request.content"]), &a)
	_ = json.Unmarshal([]byte(l[j].fields["grpc.request.content"]), &b)
	return a.Counter < b.Counter
}

func (l LogLines) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type output struct {
	m     sync.Mutex
	lines LogLines
}

func (o *output) Lines() LogLines {
	o.m.Lock()
	defer o.m.Unlock()

	retLines := make(LogLines, len(o.lines))
	copy(retLines, o.lines)

	return retLines
}

func (o *output) Append(lines ...LogLine) {
	o.m.Lock()
	defer o.m.Unlock()

	o.lines = append(o.lines, lines...)
}
func (o *output) Reset() {
	o.m.Lock()
	defer o.m.Unlock()

	o.lines = o.lines[:0]
}

type mockLogger struct {
	o *output

	fields logging.Fields
}

func newMockLogger() *mockLogger {
	return &mockLogger{o: &output{}}
}

func (l *mockLogger) Log(lvl logging.Level, msg string) {
	line := LogLine{
		lvl:    lvl,
		msg:    msg,
		fields: map[string]string{},
	}

	for i := 0; i < len(l.fields); i += 2 {
		line.fields[l.fields[i]] = l.fields[i+1]
	}
	l.o.Append(line)
}

func (l *mockLogger) With(fields ...string) logging.Logger {
	return &mockLogger{o: l.o, fields: append(append(logging.Fields{}, l.fields...), fields...)}
}

type baseLoggingSuite struct {
	*testpb.InterceptorTestSuite
	logger *mockLogger
}

func (s *baseLoggingSuite) SetupTest() {
	s.logger.fields = s.logger.fields[:0]
	s.logger.o.Reset()
}

func customClientCodeToLevel(c codes.Code) logging.Level {
	if c == codes.Unauthenticated {
		// Make this a special case for tests, and an error.
		return logging.ERROR
	}
	return logging.DefaultClientCodeToLevel(c)
}

type loggingClientServerSuite struct {
	*baseLoggingSuite
}

func TestSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}

	s := &loggingClientServerSuite{
		&baseLoggingSuite{
			logger: newMockLogger(),
			InterceptorTestSuite: &testpb.InterceptorTestSuite{
				TestService: &testpb.TestPingService{T: t},
			},
		},
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(
			addCustomFieldsUnaryClientInterceptor(),
			logging.UnaryClientInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel)),
		),
		grpc.WithChainStreamInterceptor(
			addCustomFieldsStreamClientInterceptor(),
			logging.StreamClientInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel)),
		),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.ChainStreamInterceptor(
			addCustomFieldsStreamServerInterceptor(),
			logging.StreamServerInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel))),
		grpc.ChainUnaryInterceptor(
			addCustomFieldsUnaryServerInterceptor(),
			logging.UnaryServerInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel))),
	}
	suite.Run(t, s)
}

var addCustomFields = interceptors.CommonReportableFunc(func(ctx context.Context, _ interceptors.CallMeta, _ bool) (interceptors.Reporter, context.Context) {
	// Add custom fields, one new and one that should be ignored as it duplicates the standard field.
	return interceptors.NoopReporter{}, logging.InjectFields(ctx, logging.Fields{"custom-field", "yolo", logging.ServiceFieldKey, "something different"})
})

func addCustomFieldsUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return interceptors.UnaryClientInterceptor(addCustomFields)
}

func addCustomFieldsStreamClientInterceptor() grpc.StreamClientInterceptor {
	return interceptors.StreamClientInterceptor(addCustomFields)
}

func addCustomFieldsUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(addCustomFields)
}

func addCustomFieldsStreamServerInterceptor() grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(addCustomFields)
}

func assertStandardFields(t *testing.T, kind string, f testDisposableFields, method string, typ interceptors.GRPCType) testDisposableFields {
	return f.AssertField(t, logging.SystemTag[0], logging.SystemTag[1]).
		AssertField(t, logging.ComponentFieldKey, kind).
		AssertField(t, logging.ServiceFieldKey, testpb.TestServiceFullName).
		AssertField(t, logging.MethodFieldKey, method).
		AssertField(t, logging.MethodTypeFieldKey, string(typ))
}

func (s *loggingClientServerSuite) TestPing() {
	ctx := logging.InjectFields(s.SimpleCtx(), logging.Fields{"grpc.request.value", testpb.GoodPing.Value})
	_, err := s.Client.Ping(ctx, testpb.GoodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")

	lines := s.logger.o.Lines()
	sort.Sort(lines)
	require.Len(s.T(), lines, 4)

	clientStartCallLogLine := lines[1]
	assert.Equal(s.T(), logging.DEBUG, clientStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartCallLogLine.fields, "Ping", interceptors.Unary)

	serverStartCallLogLine := lines[3]
	assert.Equal(s.T(), logging.DEBUG, serverStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartCallLogLine.fields, "Ping", interceptors.Unary)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.DEBUG, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "Ping", interceptors.Unary)
	serverFinishCallFields.AssertFieldNotEmpty(s.T(), "peer.address").
		AssertField(s.T(), "custom-field", "yolo").
		AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[0]
	assert.Equal(s.T(), logging.DEBUG, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "Ping", interceptors.Unary)
	clientFinishCallFields.AssertField(s.T(), "custom-field", "yolo").
		AssertField(s.T(), "grpc.request.value", "something").
		AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())
}

func (s *loggingClientServerSuite) TestPingList() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}
	lines := s.logger.o.Lines()
	sort.Sort(lines)
	require.Len(s.T(), lines, 4)

	serverStartCallLogLine := lines[3]
	assert.Equal(s.T(), logging.DEBUG, serverStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartCallLogLine.fields, "PingList", interceptors.ServerStream)

	clientStartCallLogLine := lines[1]
	assert.Equal(s.T(), logging.DEBUG, clientStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartCallLogLine.fields, "PingList", interceptors.ServerStream)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.DEBUG, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	serverFinishCallFields.AssertField(s.T(), "custom-field", "yolo").
		AssertFieldNotEmpty(s.T(), "peer.address").
		AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[0]
	assert.Equal(s.T(), logging.DEBUG, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	clientFinishCallFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertField(s.T(), "custom-field", "yolo").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())
}

func (s *loggingClientServerSuite) TestPingError_WithCustomLevels() {
	for _, tcase := range []struct {
		code  codes.Code
		level logging.Level
		msg   string
	}{
		{
			code:  codes.Internal,
			level: logging.WARNING,
			msg:   "Internal must remap to WarnLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.NotFound,
			level: logging.DEBUG,
			msg:   "NotFound must remap to DebugLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.FailedPrecondition,
			level: logging.DEBUG,
			msg:   "FailedPrecondition must remap to DebugLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.Unauthenticated,
			level: logging.ERROR,
			msg:   "Unauthenticated is overwritten to ErrorLevel with customClientCodeToLevel override, which probably didn't work",
		},
	} {
		s.SetupTest()
		s.T().Run(tcase.msg, func(t *testing.T) {
			_, err := s.Client.PingError(
				s.SimpleCtx(),
				&testpb.PingErrorRequest{Value: "something", ErrorCodeReturned: uint32(tcase.code)})
			require.Error(t, err, "each call here must return an error")
			lines := s.logger.o.Lines()
			sort.Sort(lines)
			require.Len(t, lines, 4)

			serverFinishCallLogLine := lines[2]
			assert.Equal(t, tcase.level, serverFinishCallLogLine.lvl)
			assert.Equal(t, "finished call", serverFinishCallLogLine.msg)
			serverFinishCallFields := assertStandardFields(t, logging.KindServerFieldValue, serverFinishCallLogLine.fields, "PingError", interceptors.Unary)
			serverFinishCallFields.AssertField(s.T(), "custom-field", "yolo").
				AssertFieldNotEmpty(t, "peer.address").
				AssertFieldNotEmpty(t, "grpc.start_time").
				AssertFieldNotEmpty(t, "grpc.request.deadline").
				AssertField(t, "grpc.code", tcase.code.String()).
				AssertField(t, "grpc.error", fmt.Sprintf("rpc error: code = %s desc = Userspace error.", tcase.code.String())).
				AssertFieldNotEmpty(t, "grpc.time_ms").AssertNoMoreTags(t)

			clientFinishCallLogLine := lines[0]
			assert.Equal(t, tcase.level, clientFinishCallLogLine.lvl)
			assert.Equal(t, "finished call", clientFinishCallLogLine.msg)
			clientFinishCallFields := assertStandardFields(t, logging.KindClientFieldValue, clientFinishCallLogLine.fields, "PingError", interceptors.Unary)
			clientFinishCallFields.AssertField(s.T(), "custom-field", "yolo").
				AssertFieldNotEmpty(t, "grpc.start_time").
				AssertFieldNotEmpty(t, "grpc.request.deadline").
				AssertField(t, "grpc.code", tcase.code.String()).
				AssertField(t, "grpc.error", fmt.Sprintf("rpc error: code = %s desc = Userspace error.", tcase.code.String())).
				AssertFieldNotEmpty(t, "grpc.time_ms").AssertNoMoreTags(t)
		})
	}
}

type loggingCustomDurationSuite struct {
	*baseLoggingSuite
}

func TestCustomDurationSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}

	s := &loggingCustomDurationSuite{
		baseLoggingSuite: &baseLoggingSuite{
			logger: newMockLogger(),
			InterceptorTestSuite: &testpb.InterceptorTestSuite{
				TestService: &testpb.TestPingService{T: t},
			},
		},
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(logging.UnaryClientInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
		grpc.WithStreamInterceptor(logging.StreamClientInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.StreamInterceptor(logging.StreamServerInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
	}
	suite.Run(t, s)
}

func (s *loggingCustomDurationSuite) TestPing_HasOverriddenDuration() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")

	lines := s.logger.o.Lines()
	sort.Sort(lines)
	require.Len(s.T(), lines, 4)

	serverStartedCallLogLine := lines[3]
	assert.Equal(s.T(), logging.INFO, serverStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartedCallLogLine.fields, "Ping", interceptors.Unary)

	clientStartedCallLogLine := lines[1]
	assert.Equal(s.T(), logging.DEBUG, clientStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartedCallLogLine.fields, "Ping", interceptors.Unary)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.INFO, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "Ping", interceptors.Unary)
	serverFinishCallFields.AssertFieldNotEmpty(s.T(), "peer.address").
		AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[0]
	assert.Equal(s.T(), logging.DEBUG, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "Ping", interceptors.Unary)
	clientFinishCallFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())
}

func (s *loggingCustomDurationSuite) TestPingList_HasOverriddenDuration() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}

	lines := s.logger.o.Lines()
	sort.Sort(lines)
	require.Len(s.T(), lines, 4)

	serverStartedCallLogLine := lines[3]
	assert.Equal(s.T(), logging.INFO, serverStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartedCallLogLine.fields, "PingList", interceptors.ServerStream)

	clientStartedCallLogLine := lines[1]
	assert.Equal(s.T(), logging.DEBUG, clientStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartedCallLogLine.fields, "PingList", interceptors.ServerStream)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.INFO, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	serverFinishCallFields.AssertFieldNotEmpty(s.T(), "peer.address").
		AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[0]
	assert.Equal(s.T(), logging.DEBUG, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	clientFinishCallFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())
}

type loggingCustomDeciderSuite struct {
	*baseLoggingSuite
}

func TestCustomDeciderSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skip("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	opts := logging.WithDecider(func(method string, _ error) logging.Decision {
		if method == path.Join("/", testpb.TestServiceFullName, "PingError") {
			return logging.LogStartAndFinishCall
		}
		return logging.NoLogCall
	})

	s := &loggingCustomDeciderSuite{
		baseLoggingSuite: &baseLoggingSuite{
			logger: newMockLogger(),
			InterceptorTestSuite: &testpb.InterceptorTestSuite{
				TestService: &testpb.TestPingService{T: t},
			},
		},
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(logging.UnaryClientInterceptor(s.logger, opts)),
		grpc.WithStreamInterceptor(logging.StreamClientInterceptor(s.logger, opts)),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.StreamInterceptor(logging.StreamServerInterceptor(s.logger, opts)),
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor(s.logger, opts)),
	}
	suite.Run(t, s)
}

func (s *loggingCustomDeciderSuite) TestPing_HasCustomDecider() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")

	require.Len(s.T(), s.logger.o.Lines(), 0) // Decider should suppress.
}

func (s *loggingCustomDeciderSuite) TestPingError_HasCustomDecider() {
	code := codes.NotFound

	_, err := s.Client.PingError(
		s.SimpleCtx(),
		&testpb.PingErrorRequest{Value: "something", ErrorCodeReturned: uint32(code)})
	require.Error(s.T(), err, "each call here must return an error")

	lines := s.logger.o.Lines()
	sort.Sort(lines)
	require.Len(s.T(), lines, 4)

	serverStartedCallLogLine := lines[3]
	assert.Equal(s.T(), logging.INFO, serverStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartedCallLogLine.fields, "PingError", interceptors.Unary)

	clientStartedCallLogLine := lines[1]
	assert.Equal(s.T(), logging.DEBUG, clientStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartedCallLogLine.fields, "PingError", interceptors.Unary)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.INFO, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "PingError", interceptors.Unary)
	serverFinishCallFields.AssertFieldNotEmpty(s.T(), "peer.address").
		AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "NotFound").
		AssertField(s.T(), "grpc.error", "rpc error: code = NotFound desc = Userspace error.").
		AssertFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[0]
	assert.Equal(s.T(), logging.DEBUG, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "PingError", interceptors.Unary)
	clientFinishCallFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "NotFound").
		AssertField(s.T(), "grpc.error", "rpc error: code = NotFound desc = Userspace error.").
		AssertFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())
}

func (s *loggingCustomDeciderSuite) TestPingList_HasCustomDecider() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}
	require.Len(s.T(), s.logger.o.Lines(), 0) // Decider should suppress.
}
