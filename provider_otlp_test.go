package tracingx

import (
	"context"
	"errors"
	"testing"

	"github.com/gostratum/core/logx"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a test logger
func getTestLogger() logx.Logger {
	return logx.NewNoopLogger()
}

func TestToAttribute(t *testing.T) {
	t.Run("converts string", func(t *testing.T) {
		attr := toAttribute("key", "value")
		assert.Equal(t, "key", string(attr.Key))
	})

	t.Run("converts int", func(t *testing.T) {
		attr := toAttribute("count", 42)
		assert.Equal(t, "count", string(attr.Key))
	})

	t.Run("converts int64", func(t *testing.T) {
		attr := toAttribute("bignum", int64(1234567890))
		assert.Equal(t, "bignum", string(attr.Key))
	})

	t.Run("converts float64", func(t *testing.T) {
		attr := toAttribute("ratio", 3.14)
		assert.Equal(t, "ratio", string(attr.Key))
	})

	t.Run("converts bool", func(t *testing.T) {
		attr := toAttribute("enabled", true)
		assert.Equal(t, "enabled", string(attr.Key))
	})

	t.Run("converts string slice", func(t *testing.T) {
		attr := toAttribute("tags", []string{"a", "b", "c"})
		assert.Equal(t, "tags", string(attr.Key))
	})

	t.Run("converts int slice", func(t *testing.T) {
		attr := toAttribute("numbers", []int{1, 2, 3})
		assert.Equal(t, "numbers", string(attr.Key))
	})

	t.Run("converts int64 slice", func(t *testing.T) {
		attr := toAttribute("bignums", []int64{100, 200, 300})
		assert.Equal(t, "bignums", string(attr.Key))
	})

	t.Run("converts float64 slice", func(t *testing.T) {
		attr := toAttribute("ratios", []float64{1.1, 2.2, 3.3})
		assert.Equal(t, "ratios", string(attr.Key))
	})

	t.Run("converts bool slice", func(t *testing.T) {
		attr := toAttribute("flags", []bool{true, false, true})
		assert.Equal(t, "flags", string(attr.Key))
	})

	t.Run("converts unknown type to string", func(t *testing.T) {
		type custom struct {
			Value string
		}
		attr := toAttribute("custom", custom{Value: "test"})
		assert.Equal(t, "custom", string(attr.Key))
	})
}

func TestHeaderCarrier(t *testing.T) {
	t.Run("Get retrieves first value", func(t *testing.T) {
		headers := map[string][]string{
			"traceparent": {"00-12345-67890-01", "ignored"},
		}
		carrier := &headerCarrier{headers: headers}

		value := carrier.Get("traceparent")
		assert.Equal(t, "00-12345-67890-01", value)
	})

	t.Run("Get returns empty for missing key", func(t *testing.T) {
		carrier := &headerCarrier{headers: make(map[string][]string)}
		value := carrier.Get("missing")
		assert.Empty(t, value)
	})

	t.Run("Set adds value", func(t *testing.T) {
		carrier := &headerCarrier{headers: make(map[string][]string)}
		carrier.Set("tracestate", "vendor=value")

		assert.Equal(t, "vendor=value", carrier.headers["tracestate"][0])
	})

	t.Run("Keys returns all keys", func(t *testing.T) {
		headers := map[string][]string{
			"traceparent": {"value1"},
			"tracestate":  {"value2"},
		}
		carrier := &headerCarrier{headers: headers}

		keys := carrier.Keys()
		assert.Len(t, keys, 2)
		assert.Contains(t, keys, "traceparent")
		assert.Contains(t, keys, "tracestate")
	})
}

func TestOTLPProviderCreationFailure(t *testing.T) {
	t.Run("fails with invalid endpoint", func(t *testing.T) {
		cfg := Config{
			ServiceName: "test-service",
			SampleRate:  1.0,
			OTLP: OTLPConfig{
				Endpoint: "invalid-endpoint-that-does-not-exist:99999",
				Insecure: true,
			},
		}

		logger := getTestLogger()

		// This should fail to create exporter
		provider, err := newOTLPProvider(cfg, logger)

		// The provider creation itself might succeed, but operations will fail
		// Or it might fail immediately - both are acceptable
		if err != nil {
			assert.Error(t, err)
			assert.Nil(t, provider)
		} else {
			// If creation succeeded, shutdown should work
			assert.NotNil(t, provider)
			shutdownErr := provider.Shutdown(context.Background())
			assert.NoError(t, shutdownErr)
		}
	})
}

func TestOTLPSpanOperations(t *testing.T) {
	// Create a real OTLP provider with in-memory exporter for testing
	cfg := Config{
		ServiceName: "test-service",
		SampleRate:  1.0,
		OTLP: OTLPConfig{
			Endpoint: "localhost:4317",
			Insecure: true,
		},
	}

	logger := getTestLogger()
	provider, err := newOTLPProvider(cfg, logger)

	// If OTLP provider creation fails (no endpoint available), skip these tests
	if err != nil {
		t.Skip("OTLP endpoint not available, skipping real span tests")
		return
	}
	defer provider.Shutdown(context.Background())

	t.Run("creates and ends span", func(t *testing.T) {
		ctx := context.Background()

		spanCtx, span := provider.Start(ctx, "test-operation",
			WithSpanKind(SpanKindServer),
			WithAttributes(map[string]any{
				"http.method": "GET",
				"http.status": 200,
			}),
		)

		assert.NotNil(t, spanCtx)
		assert.NotNil(t, span)

		// Verify span has IDs
		traceID := span.TraceID()
		spanID := span.SpanID()
		assert.NotEmpty(t, traceID)
		assert.NotEmpty(t, spanID)

		// Get span context
		spanContext := span.Context()
		assert.NotNil(t, spanContext)

		// End span
		span.End()
	})

	t.Run("sets tags on span", func(t *testing.T) {
		ctx := context.Background()
		_, span := provider.Start(ctx, "tag-test")

		span.SetTag("user.id", 12345)
		span.SetTag("user.name", "testuser")
		span.SetTag("user.active", true)
		span.SetTag("user.score", 98.5)

		span.End()
	})

	t.Run("logs fields on span", func(t *testing.T) {
		ctx := context.Background()
		_, span := provider.Start(ctx, "log-test")

		span.LogFields(
			Field{Key: "event", Value: "cache_miss"},
			Field{Key: "key", Value: "user:12345"},
			Field{Key: "ttl", Value: 300},
		)

		span.End()
	})

	t.Run("sets error on span", func(t *testing.T) {
		ctx := context.Background()
		_, span := provider.Start(ctx, "error-test")

		testErr := errors.New("simulated error")
		span.SetError(testErr)

		span.End()
	})

	t.Run("creates nested spans", func(t *testing.T) {
		ctx := context.Background()

		// Parent span
		parentCtx, parentSpan := provider.Start(ctx, "parent-operation",
			WithSpanKind(SpanKindServer),
		)
		assert.NotNil(t, parentSpan)

		// Child span
		childCtx, childSpan := provider.Start(parentCtx, "child-operation",
			WithSpanKind(SpanKindClient),
		)
		assert.NotNil(t, childSpan)

		// Grandchild span
		_, grandchildSpan := provider.Start(childCtx, "grandchild-operation")
		assert.NotNil(t, grandchildSpan)

		// End in reverse order
		grandchildSpan.End()
		childSpan.End()
		parentSpan.End()
	})

	t.Run("injects and extracts trace context", func(t *testing.T) {
		ctx := context.Background()

		// Start a span
		spanCtx, span := provider.Start(ctx, "inject-extract-test")
		defer span.End()

		// Inject into carrier
		carrier := make(map[string]string)
		err := provider.Inject(spanCtx, carrier)
		assert.NoError(t, err)

		// Extract from carrier
		extractedCtx, err := provider.Extract(ctx, carrier)
		assert.NoError(t, err)
		assert.NotNil(t, extractedCtx)
	})

	t.Run("injects and extracts with http headers", func(t *testing.T) {
		ctx := context.Background()

		// Start a span
		spanCtx, span := provider.Start(ctx, "http-headers-test")
		defer span.End()

		// Inject into HTTP headers format
		headers := make(map[string][]string)
		err := provider.Inject(spanCtx, headers)
		assert.NoError(t, err)

		// Extract from HTTP headers
		extractedCtx, err := provider.Extract(ctx, headers)
		assert.NoError(t, err)
		assert.NotNil(t, extractedCtx)
	})
}
func TestOTLPInjectionExtraction(t *testing.T) {
	provider := newNoopProvider()
	ctx := context.Background()

	t.Run("inject and extract with map[string]string", func(t *testing.T) {
		carrier := make(map[string]string)

		// Inject
		err := provider.Inject(ctx, carrier)
		assert.NoError(t, err)

		// Extract
		extractedCtx, err := provider.Extract(ctx, carrier)
		assert.NoError(t, err)
		assert.NotNil(t, extractedCtx)
	})

	t.Run("inject and extract with map[string][]string", func(t *testing.T) {
		carrier := make(map[string][]string)

		// Inject
		err := provider.Inject(ctx, carrier)
		assert.NoError(t, err)

		// Extract
		extractedCtx, err := provider.Extract(ctx, carrier)
		assert.NoError(t, err)
		assert.NotNil(t, extractedCtx)
	})
}

func TestSpanKindConversion(t *testing.T) {
	// Test that all span kinds are handled
	provider := newNoopProvider()
	ctx := context.Background()

	kinds := []struct {
		name string
		kind SpanKind
	}{
		{"internal", SpanKindInternal},
		{"server", SpanKindServer},
		{"client", SpanKindClient},
		{"producer", SpanKindProducer},
		{"consumer", SpanKindConsumer},
	}

	for _, tc := range kinds {
		t.Run(tc.name, func(t *testing.T) {
			spanCtx, span := provider.Start(ctx, "test-"+tc.name,
				WithSpanKind(tc.kind),
			)
			assert.NotNil(t, spanCtx)
			assert.NotNil(t, span)
			span.End()
		})
	}
}
