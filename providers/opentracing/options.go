package opentracing

import "github.com/opentracing/opentracing-go"

var (
	defaultOptions = &options{
		tracer: nil,
	}
)

type options struct {
	tracer          opentracing.Tracer
	traceHeaderName string
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	if optCopy.tracer == nil {
		optCopy.tracer = opentracing.GlobalTracer()
	}
	if optCopy.traceHeaderName == "" {
		optCopy.traceHeaderName = "uber-trace-id"
	}
	return optCopy
}

type Option func(*options)

// WithTraceHeaderName customizes the trace header name where trace metadata passed with requests.
// Default one is `uber-trace-id`
func WithTraceHeaderName(name string) Option {
	return func(o *options) {
		o.traceHeaderName = name
	}
}

// WithTracer sets a custom tracer to be used for this middleware, otherwise the opentracing.GlobalTracer is used.
func WithTracer(tracer opentracing.Tracer) Option {
	return func(o *options) {
		o.tracer = tracer
	}
}
