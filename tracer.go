package tracingx

import (
	"context"
	"time"
)

// Tracer provides distributed tracing capabilities
type Tracer interface {
	// Start creates a new span
	Start(ctx context.Context, operationName string, opts ...SpanOption) (context.Context, Span)

	// Extract extracts trace context from a carrier (e.g., HTTP headers)
	Extract(ctx context.Context, carrier any) (context.Context, error)

	// Inject injects trace context into a carrier (e.g., HTTP headers)
	Inject(ctx context.Context, carrier any) error

	// Shutdown gracefully shuts down the tracer
	Shutdown(ctx context.Context) error
}

// Span represents a single operation within a trace
type Span interface {
	// End completes the span
	End()

	// SetTag sets a tag/attribute on the span
	SetTag(key string, value any)

	// SetError marks the span as errored
	SetError(err error)

	// LogFields adds structured log fields to the span
	LogFields(fields ...Field)

	// Context returns the span's context
	Context() context.Context

	// TraceID returns the trace ID as a string
	TraceID() string

	// SpanID returns the span ID as a string
	SpanID() string
}

// SpanOption configures span creation
type SpanOption func(*SpanConfig)

// SpanConfig contains configuration for creating a span
type SpanConfig struct {
	Kind       SpanKind
	Attributes map[string]any
	Timestamp  time.Time
}

// SpanKind represents the type of span
type SpanKind int

const (
	// SpanKindInternal represents an internal operation
	SpanKindInternal SpanKind = iota

	// SpanKindServer represents a server-side operation
	SpanKindServer

	// SpanKindClient represents a client-side operation
	SpanKindClient

	// SpanKindProducer represents a message producer
	SpanKindProducer

	// SpanKindConsumer represents a message consumer
	SpanKindConsumer
)

// Field represents a structured log field
type Field struct {
	Key   string
	Value any
}

// WithSpanKind sets the span kind
func WithSpanKind(kind SpanKind) SpanOption {
	return func(c *SpanConfig) {
		c.Kind = kind
	}
}

// WithAttributes sets attributes on the span
func WithAttributes(attrs map[string]any) SpanOption {
	return func(c *SpanConfig) {
		if c.Attributes == nil {
			c.Attributes = make(map[string]any)
		}
		for k, v := range attrs {
			c.Attributes[k] = v
		}
	}
}

// WithTimestamp sets the span start timestamp
func WithTimestamp(t time.Time) SpanOption {
	return func(c *SpanConfig) {
		c.Timestamp = t
	}
}

// applyOptions applies span options and returns the config
func applySpanOptions(opts ...SpanOption) *SpanConfig {
	config := &SpanConfig{
		Kind:       SpanKindInternal,
		Attributes: make(map[string]any),
		Timestamp:  time.Now(),
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// Provider is the interface that tracing providers must implement
type Provider interface {
	Tracer
}

// SpanFromContext extracts a span from context
func SpanFromContext(ctx context.Context) Span {
	if span, ok := ctx.Value(spanContextKey{}).(Span); ok {
		return span
	}
	return nil
}

// ContextWithSpan returns a new context with the span attached
func ContextWithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, spanContextKey{}, span)
}

type spanContextKey struct{}
