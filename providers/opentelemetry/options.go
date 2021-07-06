package opentelemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName = "github.com/grpc-ecosystem/go-grpc-middleware/v2/providers/opentelemetry"
)

type config struct {
	TracerProvider trace.TracerProvider
	Propagators propagation.TextMapPropagator
	
	tracer trace.Tracer
}

// newConfig returns a config with all Options set.
func newConfig(opts ...Option) config {
	cfg := config{
		Propagators: otel.GetTextMapPropagator(),
		TracerProvider: otel.GetTracerProvider(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	
	cfg.tracer = cfg.TracerProvider.Tracer(instrumentationName)
	return cfg
}

// Option specifies instrumentation configuration options.
type Option func(*config)

// WithTracerProvider returns an Option to use the TracerProvider when
// creating a Tracer.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(cfg *config) {
		cfg.TracerProvider = tp
	}
}

// WithPropagators specifies propagators to use for extracting
// information from the HTTP requests. If none are specified, global
// ones will be used.
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return func(cfg *config) {
		cfg.Propagators = propagators
	}
}
