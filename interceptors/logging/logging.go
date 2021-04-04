// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package logging

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

// Decision defines rules for enabling start and end of logging.
type Decision int

const (
	// NoLogCall - Logging is disabled.
	NoLogCall Decision = iota
	// LogFinishCall - Only finish logs of request is enabled.
	LogFinishCall
	// LogStartAndFinishCall - Logging of start and end of request is enabled.
	LogStartAndFinishCall
)

var (
	// SystemTag is tag representing an event inside gRPC call.
	SystemTag = []string{"protocol", "grpc"}
	// ComponentFieldKey is a tag representing the client/server that is calling.
	ComponentFieldKey    = "grpc.component"
	KindServerFieldValue = "server"
	KindClientFieldValue = "client"
	ServiceFieldKey      = "grpc.service"
	MethodFieldKey       = "grpc.method"
	MethodTypeFieldKey   = "grpc.method_type"
)

type fieldsCtxMarker struct{}

var (
	// fieldsCtxMarkerKey is the Context value marker that is used by logging middleware to read and write logging fields into context.
	fieldsCtxMarkerKey = &fieldsCtxMarker{}
)

// Fields represents logging fields. It has to have even number of elements (pairs).
type Fields []string

func newCommonFields(kind string, c interceptors.CallMeta) Fields {
	return Fields{
		SystemTag[0], SystemTag[1],
		ComponentFieldKey, kind,
		ServiceFieldKey, c.Service,
		MethodFieldKey, c.Method,
		MethodTypeFieldKey, string(c.Typ),
	}
}

func (f Fields) Iter() FieldsIter {
	return iter{i: -2, f: f}
}

type FieldsIter interface {
	Next() bool
	At() (k, v string)
}

type iter struct {
	f Fields
	i int
}

func (i iter) Next() bool {
	if i.i >= len(i.f) {
		return false
	}

	i.i += 2
	return i.i < len(i.f)
}

func (i iter) At() (k, v string) {
	if i.i < 0 || i.i >= len(i.f) {
		return "", ""
	}

	if i.i+1 == len(i.f) {
		// Non even number of elements, add empty string.
		return i.f[i.i], ""
	}
	return i.f[i.i], i.f[i.i+1]
}

// AppendUnique returns fields which is the union of all keys with the added values having lower priority.
func (f Fields) AppendUnique(add Fields) Fields {
	if len(add) == 0 {
		return f
	}

	existing := map[string]struct{}{}
	i := f.Iter()
	for i.Next() {
		k, _ := i.At()
		existing[k] = struct{}{}
	}

	n := make(Fields, len(f), len(f)+len(add))
	copy(n, f)

	a := add.Iter()
	for a.Next() {
		k, v := a.At()
		if _, ok := existing[k]; ok {
			continue
		}
		n = append(n, k, v)
	}
	return n
}

// ExtractFields returns logging.Fields object from the Context.
// Logging interceptor adds fields into context when used.
// If no one injected fields before ExtractFields returns empty Fields.
//
// It's useful for server implementations to use this method to instantiate request logger for consistent fields (e.g request-id/tracing-id).
func ExtractFields(ctx context.Context) Fields {
	t, ok := ctx.Value(fieldsCtxMarkerKey).(Fields)
	if !ok {
		return Fields{}
	}
	n := make(Fields, len(t))
	copy(n, t)
	return n
}

// InjectFields allows to add logging.Fields that will be used logging interceptor in the path of given context (if any).
func InjectFields(ctx context.Context, f Fields) context.Context {
	return context.WithValue(ctx, fieldsCtxMarkerKey, ExtractFields(ctx).AppendUnique(f))
}

// JsonPBMarshaller is a marshaller that serializes protobuf messages.
type JsonPBMarshaller interface {
	Marshal(pb proto.Message) ([]byte, error)
}

// Logger is unified interface that we used for all our interceptors. Official implementations are available under
// provider/ directory as separate modules.
type Logger interface {
	// Log logs the fields for given log level. We can assume users (middleware library) will put fields in pairs and
	// those will be unique.
	Log(Level, string)
	// With returns Logger with given fields appended. We can assume users (middleware library) will put fields in pairs
	// and those will be unique.
	With(fields ...string) Logger
}

// Level represents logging level.
type Level string

const (
	DEBUG   = Level("debug")
	INFO    = Level("info")
	WARNING = Level("warning")
	ERROR   = Level("error")
)
