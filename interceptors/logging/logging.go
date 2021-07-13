// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logging

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func commonFields(kind string, typ interceptors.GRPCType, service string, method string) Fields {
	return Fields{
		SystemTag[0], SystemTag[1],
		ComponentFieldKey, kind,
		ServiceFieldKey, service,
		MethodFieldKey, method,
		MethodTypeFieldKey, string(typ),
	}
}

// Fields represents logging fields. It has to have even number of elements (pairs).
type Fields []string

// ErrorToCode function determines the error code of an error
// This makes using custom errors with grpc middleware easier
type ErrorToCode func(err error) codes.Code

func DefaultErrorToCode(err error) codes.Code {
	return status.Code(err)
}

// Decider function defines rules for suppressing any interceptor logs
type Decider func(fullMethodName string, err error) Decision

// DefaultDeciderMethod is the default implementation of decider to see if you should log the call
// by default this if always true so all calls are logged
func DefaultDeciderMethod(_ string, _ error) Decision {
	return LogStartAndFinishCall
}

// PayloadDecision defines rules for enabling payload logging of request and responses.
type PayloadDecision int

const (
	// NoPayloadLogging - Payload logging is disabled.
	NoPayloadLogging PayloadDecision = iota
	// LogRequest - Only logging of requests is enabled.
	LogRequest
	// LogResponse - Only logging of responses is enabled.
	LogResponse
	// LogRequestAndResponse - Logging of both requests and responses is enabled.
	LogRequestAndResponse
)

// ServerPayloadLoggingDecider is a user-provided function for deciding whether to log the server-side
// request/response payloads
type ServerPayloadLoggingDecider func(ctx context.Context, fullMethodName string, servingObject interface{}) PayloadDecision

// ClientPayloadLoggingDecider is a user-provided function for deciding whether to log the client-side
// request/response payloads
type ClientPayloadLoggingDecider func(ctx context.Context, fullMethodName string) PayloadDecision

// JsonPbMarshaler is a marshaller that serializes protobuf messages.
type JsonPbMarshaler interface {
	Marshal(pb proto.Message) ([]byte, error)
}

// Logger is unified interface that we used for all our interceptors. Official implementations are available under
// provider/ directory as separate modules.
type Logger interface {
	// Log logs the fields for given log level. We can assume users (middleware library) will put fields in pairs and
	// those will be unique.
	Log(Level, string)
	// With returns mockLogger with given fields appended. We can assume users (middleware library) will put fields in pairs
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

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level.
type CodeToLevel func(code codes.Code) Level

// DefaultServerCodeToLevel is the helper mapper that maps gRPC return codes to log levels for server side.
func DefaultServerCodeToLevel(code codes.Code) Level {
	switch code {
	case codes.OK, codes.NotFound, codes.Canceled, codes.AlreadyExists, codes.InvalidArgument, codes.Unauthenticated:
		return INFO

	case codes.DeadlineExceeded, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted,
		codes.OutOfRange, codes.Unavailable:
		return WARNING

	case codes.Unknown, codes.Unimplemented, codes.Internal, codes.DataLoss:
		return ERROR

	default:
		return ERROR
	}
}

// DefaultClientCodeToLevel is the helper mapper that maps gRPC return codes to log levels for client side.
func DefaultClientCodeToLevel(code codes.Code) Level {
	switch code {
	case codes.OK, codes.Canceled, codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.ResourceExhausted,
		codes.FailedPrecondition, codes.Aborted, codes.OutOfRange:
		return DEBUG
	case codes.Unknown, codes.DeadlineExceeded, codes.PermissionDenied, codes.Unauthenticated:
		return INFO
	case codes.Unimplemented, codes.Internal, codes.Unavailable, codes.DataLoss:
		return WARNING
	default:
		return INFO
	}
}
