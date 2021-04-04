// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package tracing_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
)

var (
	fakeInboundTraceId = "1337"
	fakeInboundSpanId  = "999"
	traceHeaderName    = "uber-trace-id"
)

type tracingAssertService struct {
	testpb.TestServiceServer
	T *testing.T
}

func (s *tracingAssertService) Ping(ctx context.Context, ping *testpb.PingRequest) (*testpb.PingResponse, error) {
	assert.NotNil(s.T, opentracing.SpanFromContext(ctx), "handlers must have the spancontext in their context, otherwise propagation will fail")
	assert.Equal(s.T, []string{tracing.FieldTraceID, fakeInboundTraceId, tracing.FieldSpanID, fakeInboundSpanId, tracing.FieldSampled, "true"}, logging.ExtractFields(ctx))
	return s.TestServiceServer.Ping(ctx, ping)
}

func (s *tracingAssertService) PingError(ctx context.Context, ping *testpb.PingErrorRequest) (*testpb.PingErrorResponse, error) {
	assert.NotNil(s.T, opentracing.SpanFromContext(ctx), "handlers must have the spancontext in their context, otherwise propagation will fail")
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *tracingAssertService) PingList(ping *testpb.PingListRequest, stream testpb.TestService_PingListServer) error {
	assert.NotNil(s.T, opentracing.SpanFromContext(stream.Context()), "handlers must have the spancontext in their context, otherwise propagation will fail")
	assert.Equal(s.T, []string{tracing.FieldTraceID, fakeInboundTraceId, tracing.FieldSpanID, fakeInboundSpanId, tracing.FieldSampled, "true"}, logging.ExtractFields(stream.Context()))
	return s.TestServiceServer.PingList(ping, stream)
}

func (s *tracingAssertService) PingEmpty(ctx context.Context, empty *testpb.PingEmptyRequest) (*testpb.PingEmptyResponse, error) {
	assert.NotNil(s.T, opentracing.SpanFromContext(ctx), "handlers must have the spancontext in their context, otherwise propagation will fail")
	assert.Equal(s.T, []string{tracing.FieldTraceID, fakeInboundTraceId, tracing.FieldSpanID, fakeInboundSpanId, tracing.FieldSampled, "true"}, logging.ExtractFields(ctx))
	return s.TestServiceServer.PingEmpty(ctx, empty)
}

func TestTracingSuite(t *testing.T) {
	mockTracer := mocktracer.New()
	opts := []tracing.Option{
		tracing.WithTracer(mockTracer),
		tracing.WithTraceHeaderName(traceHeaderName),
	}
	s := &OpentracingSuite{
		mockTracer:           mockTracer,
		InterceptorTestSuite: makeInterceptorTestSuite(t, opts),
	}
	suite.Run(t, s)
}

func TestTracingSuiteJaeger(t *testing.T) {
	mockTracer := mocktracer.New()
	mockTracer.RegisterInjector(opentracing.HTTPHeaders, jaegerFormatInjector{})
	mockTracer.RegisterExtractor(opentracing.HTTPHeaders, jaegerFormatExtractor{})
	opts := []tracing.Option{
		tracing.WithTracer(mockTracer),
	}
	s := &OpentracingSuite{
		mockTracer:           mockTracer,
		InterceptorTestSuite: makeInterceptorTestSuite(t, opts),
	}
	suite.Run(t, s)
}

func makeInterceptorTestSuite(t *testing.T, opts []tracing.Option) *testpb.InterceptorTestSuite {
	return &testpb.InterceptorTestSuite{
		TestService: &tracingAssertService{TestServiceServer: &testpb.TestPingService{T: t}, T: t},
		ClientOpts: []grpc.DialOption{
			grpc.WithUnaryInterceptor(tracing.UnaryClientInterceptor(opts...)),
			grpc.WithStreamInterceptor(tracing.StreamClientInterceptor(opts...)),
		},
		ServerOpts: []grpc.ServerOption{
			grpc.StreamInterceptor(tracing.StreamServerInterceptor(opts...)),
			grpc.UnaryInterceptor(tracing.UnaryServerInterceptor(opts...)),
		},
	}
}

type OpentracingSuite struct {
	*testpb.InterceptorTestSuite
	mockTracer *mocktracer.MockTracer
}

func (s *OpentracingSuite) SetupTest() {
	s.mockTracer.Reset()
}

func (s *OpentracingSuite) createContextFromFakeHttpRequestParent(ctx context.Context, sampled bool) context.Context {
	jFlag := 0
	if sampled {
		jFlag = 1
	}

	hdr := http.Header{}
	hdr.Set(traceHeaderName, fmt.Sprintf("%s:%s:%s:%d", fakeInboundTraceId, fakeInboundSpanId, fakeInboundSpanId, jFlag))
	hdr.Set("mockpfx-ids-traceid", fakeInboundTraceId)
	hdr.Set("mockpfx-ids-spanid", fakeInboundSpanId)
	hdr.Set("mockpfx-ids-sampled", fmt.Sprint(sampled))

	parentSpanContext, err := s.mockTracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdr))
	require.NoError(s.T(), err, "parsing a fake HTTP request headers shouldn't fail, ever")
	fakeSpan := s.mockTracer.StartSpan(
		"/fake/parent/http/request",
		// this is magical, it attaches the new span to the parent parentSpanContext, and creates an unparented one if empty.
		opentracing.ChildOf(parentSpanContext),
	)
	fakeSpan.Finish()
	return opentracing.ContextWithSpan(ctx, fakeSpan)
}

func (s *OpentracingSuite) assertTracesCreated(methodName string) (clientSpan *mocktracer.MockSpan, serverSpan *mocktracer.MockSpan) {
	spans := s.mockTracer.FinishedSpans()
	for _, span := range spans {
		s.T().Logf("span: %v, tags: %v", span, span.Tags())
	}
	require.Len(s.T(), spans, 3, "should record 3 spans: one fake inbound, one client, one server")
	traceIdAssert := fmt.Sprintf("traceId=%s", fakeInboundTraceId)
	for _, span := range spans {
		assert.Contains(s.T(), span.String(), traceIdAssert, "not part of the fake parent trace: %v", span)
		if span.OperationName == methodName {
			kind := fmt.Sprintf("%v", span.Tag("span.kind"))
			if kind == "client" {
				clientSpan = span
			} else if kind == "server" {
				serverSpan = span
			}
			assert.EqualValues(s.T(), span.Tag("component"), "gRPC", "span must be tagged with gRPC component")
		}
	}
	require.NotNil(s.T(), clientSpan, "client span must be there")
	require.NotNil(s.T(), serverSpan, "server span must be there")
	assert.EqualValues(s.T(), serverSpan.Tag("grpc.request.value"), "something", "tags must be propagated, in this case ones from request fields")
	return clientSpan, serverSpan
}

func (s *OpentracingSuite) TestPing_PropagatesTraces() {
	ctx := s.createContextFromFakeHttpRequestParent(s.SimpleCtx(), true)
	_, err := s.Client.Ping(ctx, testpb.GoodPing)
	require.NoError(s.T(), err, "there must be not be an on a successful call")
	s.assertTracesCreated("/" + testpb.TestServiceFullName + "/Ping")
}

func (s *OpentracingSuite) TestPing_ClientContextTags() {
	const name = "opentracing.custom"
	ctx := tracing.ClientAddContextTags(
		s.createContextFromFakeHttpRequestParent(s.SimpleCtx(), true),
		opentracing.Tags{name: ""},
	)

	_, err := s.Client.Ping(ctx, testpb.GoodPing)
	require.NoError(s.T(), err, "there must be not be an on a successful call")

	for _, span := range s.mockTracer.FinishedSpans() {
		if span.OperationName == "/"+testpb.TestServiceFullName+"/Ping" {
			kind := fmt.Sprintf("%v", span.Tag("span.kind"))
			if kind == "client" {
				assert.Contains(s.T(), span.Tags(), name, "custom opentracing.Tags must be included in context")
			}
		}
	}
}

func (s *OpentracingSuite) TestPingList_PropagatesTraces() {
	ctx := s.createContextFromFakeHttpRequestParent(s.SimpleCtx(), true)
	stream, err := s.Client.PingList(ctx, testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}
	s.assertTracesCreated("/" + testpb.TestServiceFullName + "/PingList")
}

func (s *OpentracingSuite) TestPingError_PropagatesTraces() {
	ctx := s.createContextFromFakeHttpRequestParent(s.SimpleCtx(), true)
	_, err := s.Client.PingError(ctx, &testpb.PingErrorRequest{Value: "something", ErrorCodeReturned: uint32(codes.OutOfRange)})
	require.Error(s.T(), err, "there must be an error returned here")
	clientSpan, serverSpan := s.assertTracesCreated("/" + testpb.TestServiceFullName + "/PingError")
	assert.Equal(s.T(), true, clientSpan.Tag("error"), "client span needs to be marked as an error")
	assert.Equal(s.T(), true, serverSpan.Tag("error"), "server span needs to be marked as an error")
}

func (s *OpentracingSuite) TestPingEmpty_NotSampleTraces() {
	ctx := s.createContextFromFakeHttpRequestParent(s.SimpleCtx(), false)
	_, err := s.Client.PingEmpty(ctx, &testpb.PingEmptyRequest{})
	require.NoError(s.T(), err, "there must be not be an on a successful call")
}

type jaegerFormatInjector struct{}

func (jaegerFormatInjector) Inject(ctx mocktracer.MockSpanContext, carrier interface{}) error {
	w := carrier.(opentracing.TextMapWriter)
	flags := 0
	if ctx.Sampled {
		flags = 1
	}
	w.Set(traceHeaderName, fmt.Sprintf("%d:%d::%d", ctx.TraceID, ctx.SpanID, flags))

	return nil
}

type jaegerFormatExtractor struct{}

func (jaegerFormatExtractor) Extract(carrier interface{}) (mocktracer.MockSpanContext, error) {
	rval := mocktracer.MockSpanContext{Sampled: true}
	reader, ok := carrier.(opentracing.TextMapReader)
	if !ok {
		return rval, opentracing.ErrInvalidCarrier
	}
	err := reader.ForeachKey(func(key, val string) error {
		lowerKey := strings.ToLower(key)
		switch {
		case lowerKey == traceHeaderName:
			parts := strings.Split(val, ":")
			if len(parts) != 4 {
				return errors.New("invalid trace id format")
			}
			traceId, err := strconv.Atoi(parts[0])
			if err != nil {
				return err
			}
			rval.TraceID = traceId
			spanId, err := strconv.Atoi(parts[1])
			if err != nil {
				return err
			}
			rval.SpanID = spanId
			flags, err := strconv.Atoi(parts[3])
			if err != nil {
				return err
			}
			rval.Sampled = flags%2 == 1
		}
		return nil
	})
	if rval.TraceID == 0 || rval.SpanID == 0 {
		return rval, opentracing.ErrSpanContextNotFound
	}
	if err != nil {
		return rval, err
	}
	return rval, nil
}
