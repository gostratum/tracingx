package tracingx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoopProvider(t *testing.T) {
	provider := newNoopProvider()

	t.Run("Start creates span", func(t *testing.T) {
		ctx := context.Background()
		spanCtx, span := provider.Start(ctx, "test-operation")

		assert.NotNil(t, spanCtx)
		assert.NotNil(t, span)

		// Verify span is in context
		retrievedSpan := SpanFromContext(spanCtx)
		assert.Equal(t, span, retrievedSpan)
	})

	t.Run("Start with options", func(t *testing.T) {
		ctx := context.Background()
		attrs := map[string]interface{}{
			"http.method": "POST",
		}

		spanCtx, span := provider.Start(ctx, "http-request",
			WithSpanKind(SpanKindServer),
			WithAttributes(attrs),
		)

		assert.NotNil(t, spanCtx)
		assert.NotNil(t, span)
	})

	t.Run("Extract returns context unchanged", func(t *testing.T) {
		ctx := context.Background()
		carrier := map[string]string{
			"traceparent": "00-12345-67890-01",
		}

		resultCtx, err := provider.Extract(ctx, carrier)
		assert.NoError(t, err)
		assert.Equal(t, ctx, resultCtx)
	})

	t.Run("Inject does nothing", func(t *testing.T) {
		ctx := context.Background()
		carrier := make(map[string]string)

		err := provider.Inject(ctx, carrier)
		assert.NoError(t, err)
		assert.Empty(t, carrier)
	})

	t.Run("Shutdown does nothing", func(t *testing.T) {
		ctx := context.Background()
		err := provider.Shutdown(ctx)
		assert.NoError(t, err)
	})
}

func TestNoopSpan(t *testing.T) {
	provider := newNoopProvider()
	ctx := context.Background()
	_, span := provider.Start(ctx, "test")

	t.Run("End does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			span.End()
		})
	})

	t.Run("SetTag does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			span.SetTag("key", "value")
			span.SetTag("number", 123)
			span.SetTag("bool", true)
		})
	})

	t.Run("SetError does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			span.SetError(errors.New("test error"))
			span.SetError(nil)
		})
	})

	t.Run("LogFields does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			span.LogFields(
				Field{Key: "event", Value: "cache_miss"},
				Field{Key: "key", Value: "user:123"},
			)
		})
	})

	t.Run("Context returns context", func(t *testing.T) {
		spanCtx := span.Context()
		assert.NotNil(t, spanCtx)
	})

	t.Run("TraceID returns empty string", func(t *testing.T) {
		traceID := span.TraceID()
		assert.Empty(t, traceID)
	})

	t.Run("SpanID returns empty string", func(t *testing.T) {
		spanID := span.SpanID()
		assert.Empty(t, spanID)
	})
}

func TestNoopSpanCompleteWorkflow(t *testing.T) {
	t.Run("complete tracing workflow with noop", func(t *testing.T) {
		provider := newNoopProvider()
		ctx := context.Background()

		// Start parent span
		parentCtx, parentSpan := provider.Start(ctx, "parent-operation",
			WithSpanKind(SpanKindServer),
			WithAttributes(map[string]interface{}{
				"http.method": "GET",
				"http.path":   "/api/users",
			}),
		)
		assert.NotNil(t, parentSpan)

		// Simulate work and log
		parentSpan.SetTag("db.query", "SELECT * FROM users")
		parentSpan.LogFields(Field{Key: "event", Value: "query_start"})

		// Start child span from parent context
		childCtx, childSpan := provider.Start(parentCtx, "database-query",
			WithSpanKind(SpanKindClient),
		)
		assert.NotNil(t, childSpan)

		// Simulate database work
		childSpan.SetTag("db.rows", 42)
		childSpan.LogFields(Field{Key: "event", Value: "query_complete"})
		childSpan.End()

		// Complete parent
		parentSpan.LogFields(Field{Key: "event", Value: "request_complete"})
		parentSpan.End()

		// Extract/inject for distributed tracing
		carrier := make(map[string]string)
		err := provider.Inject(childCtx, carrier)
		assert.NoError(t, err)

		newCtx, err := provider.Extract(ctx, carrier)
		assert.NoError(t, err)
		assert.NotNil(t, newCtx)

		// Shutdown
		err = provider.Shutdown(ctx)
		assert.NoError(t, err)
	})
}
