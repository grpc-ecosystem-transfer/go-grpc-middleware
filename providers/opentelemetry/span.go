package opentelemetry

import (
	"context"
	
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	grpccodes "google.golang.org/grpc/codes"
	
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing/kv"
)

type span struct {
	span trace.Span
	ctx  context.Context
}

func newSpan(ctx context.Context, rawSpan trace.Span) *span {
	return &span{
		ctx:  ctx,
		span: rawSpan,
	}
}

func (s *span) End() {
	s.span.End()
}

func (s *span) SetStatus(code grpccodes.Code, message string) {
	if code != grpccodes.OK {
		s.span.SetStatus(codes.Error, message)
	} else {
		s.span.SetStatus(codes.Ok, message)
	}
}

func (s *span) SetAttributes(attrs ...kv.KeyValue) {
	s.span.SetAttributes(translateKeyValue(attrs)...)
}

func (s *span) AddEvent(name string, attrs ...kv.KeyValue) {
	kvList := translateKeyValue(attrs)
	
	s.span.AddEvent(name, trace.WithAttributes(kvList...))
}

func translateKeyValue(kvs []kv.KeyValue) []attribute.KeyValue {
	kvList := make([]attribute.KeyValue, 0, len(kvs))
	for _, v := range kvs {
		var otelKeyValue attribute.KeyValue
		otelKey := attribute.Key(v.Key)
		
		switch v.Value.Type() {
		case kv.BOOL:
			otelKeyValue = otelKey.Bool(v.Value.AsBool())
		case kv.INT64:
			otelKeyValue = otelKey.Int64(v.Value.AsInt64())
		case kv.FLOAT64:
			otelKeyValue = otelKey.Float64(v.Value.AsFloat64())
		case kv.STRING:
			otelKeyValue = otelKey.String(v.Value.AsString())
		default:
			continue
		}
		kvList = append(kvList, otelKeyValue)
	}
	return kvList
}
