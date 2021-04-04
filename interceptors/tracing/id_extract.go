package tracing

import (
	"strings"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/grpclog"
)

const (
	FieldTraceID         = "trace.traceid"
	FieldSpanID          = "trace.spanid"
	FieldSampled         = "trace.sampled"
	jaegerNotSampledFlag = "0"
)

// getTraceMeta returns trace hidden data from tracer.
// This is done in an incredibly hacky way, because the public-facing interface of opentracing doesn't give access to
// the TraceId and SpanId of the SpanContext. Only the Tracer's Inject/Extract methods know what these are.
// Most tracers have them encoded as keys with 'traceid' and 'spanid':
// https://github.com/openzipkin/zipkin-go-opentracing/blob/594640b9ef7e5c994e8d9499359d693c032d738c/propagation_ot.go#L29
// https://github.com/opentracing/basictracer-go/blob/1b32af207119a14b1b231d451df3ed04a72efebf/propagation_ot.go#L26
// Jaeger from Uber use one-key schema with next format '{trace-id}:{span-id}:{parent-span-id}:{flags}'
// https://www.jaegertracing.io/docs/client-libraries/#trace-span-identity
// Datadog uses keys ending with 'trace-id' and 'parent-id' (for span) by default:
// https://github.com/DataDog/dd-trace-go/blob/v1/ddtrace/tracer/textmap.go#L77
func getTraceMeta(traceHeaderName string, span opentracing.Span) TraceMeta {
	c := &mockedCarrier{traceHeaderName: traceHeaderName}
	if err := span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, c); err != nil {
		grpclog.Infof("grpc_opentracing: failed extracting trace info into ctx %v", err)
	}
	return c.m
}

type TraceMeta struct {
	TraceID string
	SpanID  string
	Sampled bool
}

type mockedCarrier struct {
	m               TraceMeta
	traceHeaderName string
}

func (c *mockedCarrier) Set(key, val string) {
	key = strings.ToLower(key)

	if key == c.traceHeaderName {
		parts := strings.Split(val, ":")
		if len(parts) == 4 {
			c.m.TraceID = parts[0]
			c.m.SpanID = parts[1]

			c.m.Sampled = parts[3] != jaegerNotSampledFlag
			return
		}
	}

	if strings.Contains(key, "traceid") {
		c.m.TraceID = val // This will most likely be base-16 (hex) encoded.
	}

	if strings.Contains(key, "spanid") && !strings.Contains(strings.ToLower(key), "parent") {
		c.m.SpanID = val // This will most likely be base-16 (hex) encoded.
	}

	if strings.Contains(key, "sampled") {
		c.m.Sampled = val == "true"
	}

	if strings.HasSuffix(key, "trace-id") {
		c.m.TraceID = val
	}

	if strings.HasSuffix(key, "parent-id") {
		c.m.SpanID = val
	}
}
