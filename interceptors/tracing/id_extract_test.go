package tracing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockedCarrier_Set_JaegerTraceFormat(t *testing.T) {
	var (
		fakeTraceSampled   = 1
		fakeInboundTraceId = "deadbeef"
		fakeInboundSpanId  = "c0decafe"
		traceHeaderName    = "uber-trace-id"
	)

	traceHeaderValue := fmt.Sprintf("%s:%s:%s:%d", fakeInboundTraceId, fakeInboundSpanId, fakeInboundSpanId, fakeTraceSampled)

	c := &mockedCarrier{traceHeaderName: traceHeaderName}
	c.Set(traceHeaderName, traceHeaderValue)
	assert.Equal(t, TraceMeta{
		TraceID: fakeInboundTraceId,
		SpanID:  fakeInboundSpanId,
		Sampled: true,
	}, c.m)
}
