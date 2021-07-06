package opentracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/codes"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing/kv"
)

type span struct {
	span opentracing.Span
	ctx  context.Context

	initial bool
}

// Compatibility check.
var _ tracing.Span = &span{}

func newSpan(rawSpan opentracing.Span, ctx context.Context) *span {
	return &span{
		span: rawSpan,
		ctx:     ctx,
		initial: true,
	}
}

func (s *span) End() {
	// Middleware tags only record once
	if s.initial {
		s.initial = false
		t := tags.Extract(s.ctx)
		for k, v := range t.Values() {
			s.span.SetTag(k, v)
		}
	}

	s.span.Finish()
}

func (s *span) SetStatus(code codes.Code, msg string) {
	if code != codes.OK {
		ext.Error.Set(s.span, true)
		s.span.LogFields(log.String("event", "error"), log.String("message", msg))
	}
}

func (s *span) AddEvent(name string, attrs ...kv.KeyValue) {
	fields := make([]log.Field, 0, len(attrs) +1)
	
	fields = append(fields, log.String("event", name))
	
	for _, attr := range attrs {
		switch attr.Value.Type() {
		case kv.BOOL:
			fields = append(fields, log.Bool(string(attr.Key), attr.Value.AsBool()))
		case kv.INT64:
			fields = append(fields, log.Int64(string(attr.Key), attr.Value.AsInt64()))
		case kv.FLOAT64:
			fields = append(fields, log.Float64(string(attr.Key), attr.Value.AsFloat64()))
		case kv.STRING:
			fields = append(fields, log.String(string(attr.Key), attr.Value.AsString()))
		default:
			continue
		}
	}
	
	s.span.LogFields(fields...)
}


func (s *span) SetAttributes(attrs ...kv.KeyValue) {
	for _, attr := range attrs {
		var v interface{}
		switch attr.Value.Type() {
		case kv.BOOL:
			v = attr.Value.AsBool()
		case kv.INT64:
			v = attr.Value.AsInt64()
		case kv.FLOAT64:
			v = attr.Value.AsFloat64()
		case kv.STRING:
			v = attr.Value.AsString()
		default:
			continue
		}
		
		if v != nil {
			s.span.SetTag(string(attr.Key), v)
		}
	}
}

