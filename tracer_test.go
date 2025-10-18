package tracingx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSpanOptions(t *testing.T) {
	t.Run("WithSpanKind", func(t *testing.T) {
		opt := WithSpanKind(SpanKindServer)
		config := &SpanConfig{}
		opt(config)
		assert.Equal(t, SpanKindServer, config.Kind)
	})

	t.Run("WithAttributes", func(t *testing.T) {
		attrs := map[string]any{
			"http.method": "GET",
			"http.url":    "/api/users",
		}
		opt := WithAttributes(attrs)
		config := &SpanConfig{
			Attributes: make(map[string]any),
		}
		opt(config)
		assert.Equal(t, "GET", config.Attributes["http.method"])
		assert.Equal(t, "/api/users", config.Attributes["http.url"])
	})

	t.Run("WithTimestamp", func(t *testing.T) {
		timestamp := time.Now().Add(-1 * time.Hour)
		opt := WithTimestamp(timestamp)
		config := &SpanConfig{}
		opt(config)
		assert.Equal(t, timestamp, config.Timestamp)
	})
}

func TestApplySpanOptions(t *testing.T) {
	t.Run("applies multiple options", func(t *testing.T) {
		timestamp := time.Now().Add(-1 * time.Hour)
		attrs := map[string]any{
			"service": "user-api",
		}

		config := applySpanOptions(
			WithSpanKind(SpanKindClient),
			WithAttributes(attrs),
			WithTimestamp(timestamp),
		)

		assert.Equal(t, SpanKindClient, config.Kind)
		assert.Equal(t, "user-api", config.Attributes["service"])
		assert.Equal(t, timestamp, config.Timestamp)
	})

	t.Run("applies no options with defaults", func(t *testing.T) {
		config := applySpanOptions()

		assert.Equal(t, SpanKindInternal, config.Kind)
		assert.NotNil(t, config.Attributes)
		assert.False(t, config.Timestamp.IsZero())
	})
}

func TestSpanContext(t *testing.T) {
	t.Run("ContextWithSpan and SpanFromContext", func(t *testing.T) {
		ctx := context.Background()
		provider := newNoopProvider()

		// Create span
		spanCtx, span := provider.Start(ctx, "test-operation")

		// Verify span is in context
		retrievedSpan := SpanFromContext(spanCtx)
		assert.NotNil(t, retrievedSpan)
		assert.Equal(t, span, retrievedSpan)
	})

	t.Run("SpanFromContext returns nil when no span", func(t *testing.T) {
		ctx := context.Background()
		span := SpanFromContext(ctx)
		assert.Nil(t, span)
	})

	t.Run("ContextWithSpan creates new context", func(t *testing.T) {
		ctx := context.Background()
		provider := newNoopProvider()
		_, span := provider.Start(ctx, "test")

		newCtx := ContextWithSpan(ctx, span)
		assert.NotNil(t, newCtx)

		retrievedSpan := SpanFromContext(newCtx)
		assert.Equal(t, span, retrievedSpan)
	})
}

func TestSpanKinds(t *testing.T) {
	t.Run("all span kinds are unique", func(t *testing.T) {
		kinds := []SpanKind{
			SpanKindInternal,
			SpanKindServer,
			SpanKindClient,
			SpanKindProducer,
			SpanKindConsumer,
		}

		seen := make(map[SpanKind]bool)
		for _, kind := range kinds {
			assert.False(t, seen[kind], "duplicate span kind: %v", kind)
			seen[kind] = true
		}
	})
}

func TestField(t *testing.T) {
	t.Run("creates field with key and value", func(t *testing.T) {
		field := Field{
			Key:   "user_id",
			Value: 12345,
		}

		assert.Equal(t, "user_id", field.Key)
		assert.Equal(t, 12345, field.Value)
	})
}
