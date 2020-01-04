package ctxzr

import (
	"context"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/rs/zerolog"
)

type ctxMarker struct{}

type CtxLogger struct {
	Logger *zerolog.Logger
	Fields map[string]interface{}
}

var (
	ctxMarkerKey = &ctxMarker{}
	nullLogger   = &zerolog.Logger{}
)

func MergeFields(mp1 map[string]interface{}, mp2 map[string]interface{}) map[string]interface{} {

	mp3 := make(map[string]interface{}, 0)
	for k, v := range mp1 {
		if _, ok := mp1[k]; ok {
			mp3[k] = v
		}
	}

	for k, v := range mp2 {
		if _, ok := mp2[k]; ok {
			mp3[k] = v
		}
	}
	return mp3
}

// AddFields adds fields to the logger.
func AddFields(ctx context.Context, fields map[string]interface{}) {
	l, ok := ctx.Value(ctxMarkerKey).(*CtxLogger)
	if !ok || l == nil {
		return
	}

	for k, v := range fields {
		l.Fields[k] = v
	}

}

// Extract takes the call-scoped Logger from grpc middleware.
//
// It always returns a Logger that has all the grpc_ctxtags updated.
func Extract(ctx context.Context) *CtxLogger {
	l, ok := ctx.Value(ctxMarkerKey).(*CtxLogger)
	if !ok || l == nil {
		return &CtxLogger{Logger: nullLogger, Fields: make(map[string]interface{}, 0)}
	}
	// Add grpc_ctxtags tags metadata until now.
	fields := TagsToFields(ctx)
	// Addfields added until now.
	fields = MergeFields(fields, l.Fields)
	return &CtxLogger{Logger: l.Logger, Fields: fields}
}

// TagsToFields transforms the Tags on the supplied context into zap fields.
func TagsToFields(ctx context.Context) map[string]interface{} {
	fields := make(map[string]interface{}, 0)
	tags := grpc_ctxtags.Extract(ctx)
	for k, v := range tags.Values() {
		fields[k] = v
	}
	return fields
}

// ToContext adds the zerolog.Logger to the context for extraction later.
// Returning the new context that has been created.
func ToContext(ctx context.Context, logger *CtxLogger) context.Context {
	return context.WithValue(ctx, ctxMarkerKey, logger)
}
