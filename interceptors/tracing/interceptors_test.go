package tracing_test

import (
	"context"
	"io"
	"strconv"
	"sync/atomic"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing/kv"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
)

var (
	id               int64 = 0
	traceIDHeaderKey       = "traceid"
	spanIDHeaderKey        = "spanid"
)

func extractFromContext(ctx context.Context, kind tracing.SpanKind) *mockSpan {
	var m metadata.MD
	if kind == tracing.SpanKindClient {
		m, _ = metadata.FromOutgoingContext(ctx)
	} else {
		m, _ = metadata.FromIncomingContext(ctx)
	}
	
	traceIDValues := m.Get(traceIDHeaderKey)
	if len(traceIDValues) == 0 {
		return nil
	}
	spanIDValues := m.Get(spanIDHeaderKey)
	if len(spanIDValues) == 0 {
		return nil
	}
	
	return &mockSpan{
		traceID: traceIDValues[0],
		spanID:  spanIDValues[0],
	}
}

func injectWithContext(ctx context.Context, span *mockSpan, kind tracing.SpanKind) context.Context {
	var m metadata.MD
	if kind == tracing.SpanKindClient {
		m, _ = metadata.FromOutgoingContext(ctx)
	} else {
		m, _ = metadata.FromIncomingContext(ctx)
	}
	m = m.Copy()
	
	m.Set(traceIDHeaderKey, span.traceID)
	m.Set(spanIDHeaderKey, span.spanID)
	
	ctx = metadata.NewOutgoingContext(ctx, m)
	return ctx
}

func genID() string {
	return strconv.FormatInt(atomic.AddInt64(&id, 1), 10)
}

// Implements Tracker
type mockTracer struct {
	spanStore map[string]*mockSpan
}

func (t *mockTracer) ListSpan(kind tracing.SpanKind) []*mockSpan {
	var spans []*mockSpan
	for _, v := range t.spanStore {
		if v.kind == kind {
			spans = append(spans, v)
		}
	}
	return spans
}

func (t *mockTracer) Reset() {
	t.spanStore = make(map[string]*mockSpan)
}

func newMockTracer() *mockTracer {
	return &mockTracer{
		spanStore: make(map[string]*mockSpan),
	}
}

func (t *mockTracer) Start(ctx context.Context, spanName string, kind tracing.SpanKind) (context.Context, tracing.Span) {
	span := mockSpan{
		spanID:     genID(),
		name:       spanName,
		kind:       kind,
		statusCode: codes.OK,
	}
	
	// parentSpan := spanFromContext(ctx)
	parentSpan := extractFromContext(ctx, kind)
	if parentSpan != nil {
		// Fetch span from context as parent span
		span.traceID = parentSpan.traceID
		span.parentSpanID = parentSpan.spanID
	} else {
		span.traceID = genID()
	}
	
	t.spanStore[span.spanID] = &span
	
	// ctx = contextWithSpan(ctx, &span)
	if kind == tracing.SpanKindClient {
		ctx = injectWithContext(ctx, &span, kind)
	}
	return ctx, &span
}

// Implements Span
type mockSpan struct {
	traceID      string
	spanID       string
	parentSpanID string
	
	name string
	kind tracing.SpanKind
	end  bool
	
	statusCode    codes.Code
	statusMessage string
	
	msgSendCounter     int
	msgReceivedCounter int
	eventNameList      []string
	attributesList     [][]kv.KeyValue
}

func (s *mockSpan) SetAttributes(attrs ...kv.KeyValue) {
	s.attributesList = append(s.attributesList, attrs)
}

func (s *mockSpan) End() {
	s.end = true
}

func (s *mockSpan) SetStatus(code codes.Code, message string) {
	s.statusCode = code
	s.statusMessage = message
}

func (s *mockSpan) AddEvent(name string, attrs ...kv.KeyValue) {
	s.eventNameList = append(s.eventNameList, name)
	
	for _, v := range attrs {
		switch v {
		case tracing.RPCMessageTypeSent:
			s.msgSendCounter++
		case tracing.RPCMessageTypeReceived:
			s.msgReceivedCounter++
		}
	}
}

type tracingSuite struct {
	*testpb.InterceptorTestSuite
	tracer *mockTracer
}

func (s *tracingSuite) BeforeTest(suiteName, testName string) {
	s.tracer.Reset()
}

func (s *tracingSuite) TestPing() {
	method := "/testing.testpb.v1.TestService/Ping"
	errorMethod := "/testing.testpb.v1.TestService/PingError"
	t := s.T()
	
	testCases := []struct {
		name         string
		error        bool
		errorMessage string
	}{
		{
			name:  "OK",
			error: false,
		},
		{
			name:         "invalid argument error",
			error:        true,
			errorMessage: "Userspace error.",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s.tracer.Reset()
			
			var err error
			if tc.error {
				req := &testpb.PingErrorRequest{ErrorCodeReturned: uint32(codes.InvalidArgument)}
				_, err = s.Client.PingError(s.SimpleCtx(), req)
			} else {
				req := &testpb.PingRequest{Value: "something"}
				_, err = s.Client.Ping(s.SimpleCtx(), req)
			}
			if tc.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			
			clientSpans := s.tracer.ListSpan(tracing.SpanKindClient)
			serverSpans := s.tracer.ListSpan(tracing.SpanKindServer)
			require.Len(t, clientSpans, 1)
			require.Len(t, serverSpans, 1)
			
			clientSpan := clientSpans[0]
			assert.True(t, clientSpan.end)
			assert.Equal(t, 1, clientSpan.msgSendCounter)
			assert.Equal(t, 1, clientSpan.msgReceivedCounter)
			assert.Equal(t, []string{"message", "message"}, clientSpan.eventNameList)
			
			serverSpan := serverSpans[0]
			assert.True(t, serverSpan.end)
			assert.Equal(t, 1, serverSpan.msgSendCounter)
			assert.Equal(t, 1, serverSpan.msgReceivedCounter)
			assert.Equal(t, []string{"message", "message"}, serverSpan.eventNameList)
			
			assert.Equal(t, clientSpan.traceID, serverSpan.traceID)
			assert.Equal(t, clientSpan.spanID, serverSpan.parentSpanID)
			
			if tc.error {
				assert.Equal(t, codes.InvalidArgument, clientSpan.statusCode)
				assert.Equal(t, tc.errorMessage, clientSpan.statusMessage)
				assert.Equal(t, errorMethod, clientSpan.name)
				assert.Equal(t, [][]kv.KeyValue{{kv.Key("rpc.grpc.status_code").Int64(3)}}, clientSpan.attributesList)
				
				assert.Equal(t, errorMethod, serverSpan.name)
				assert.Equal(t, [][]kv.KeyValue{{kv.Key("rpc.grpc.status_code").Int64(3)}}, serverSpan.attributesList)
			} else {
				assert.Equal(t, codes.OK, clientSpan.statusCode)
				assert.Equal(t, method, clientSpan.name)
				assert.Equal(t, [][]kv.KeyValue{{kv.Key("rpc.grpc.status_code").Int64(0)}}, clientSpan.attributesList)
				
				assert.Equal(t, method, serverSpan.name)
				assert.Equal(t, [][]kv.KeyValue{{kv.Key("rpc.grpc.status_code").Int64(0)}}, serverSpan.attributesList)
			}
		})
	}
}

func (s *tracingSuite) TestPingList() {
	t := s.T()
	method := "/testing.testpb.v1.TestService/PingList"
	
	stream, err := s.Client.PingList(s.SimpleCtx(), &testpb.PingListRequest{Value: "something"})
	require.NoError(t, err)
	
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}
	
	clientSpans := s.tracer.ListSpan(tracing.SpanKindClient)
	serverSpans := s.tracer.ListSpan(tracing.SpanKindServer)
	require.Len(t, clientSpans, 1)
	require.Len(t, serverSpans, 1)
	
	clientSpan := clientSpans[0]
	assert.True(t, clientSpan.end)
	assert.Equal(t, 1, clientSpan.msgSendCounter)
	assert.Equal(t, testpb.ListResponseCount+1, clientSpan.msgReceivedCounter)
	assert.Equal(t, codes.OK, clientSpan.statusCode)
	assert.Equal(t, method, clientSpan.name)
	
	serverSpan := serverSpans[0]
	assert.True(t, serverSpan.end)
	assert.Equal(t, testpb.ListResponseCount, serverSpan.msgSendCounter)
	assert.Equal(t, 1, serverSpan.msgReceivedCounter)
	assert.Equal(t, codes.OK, serverSpan.statusCode)
	assert.Equal(t, method, serverSpan.name)
}

func TestSuite(t *testing.T) {
	tracer := newMockTracer()
	
	s := tracingSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: &testpb.TestPingService{T: t},
		},
		tracer: tracer,
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(tracing.UnaryClientInterceptor(tracer)),
		grpc.WithStreamInterceptor(tracing.StreamClientInterceptor(tracer)),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			tags.UnaryServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			tracing.UnaryServerInterceptor(tracer),
		),
		grpc.ChainStreamInterceptor(
			tags.StreamServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			tracing.StreamServerInterceptor(tracer),
		),
	}
	
	suite.Run(t, &s)
}
