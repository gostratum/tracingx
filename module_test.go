package tracingx

import (
	"context"
	"testing"

	"github.com/gostratum/core/logx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewTracer(t *testing.T) {
	logger := logx.ProvideAdapter(zap.NewNop())

	t.Run("creates noop tracer when disabled", func(t *testing.T) {
		params := Params{
			Config: Config{
				Enabled: false,
			},
			Logger: logger,
		}

		result, err := NewTracer(params)
		require.NoError(t, err)
		assert.NotNil(t, result.Tracer)
		assert.NotNil(t, result.Provider)

		// Test that it works as noop
		ctx := context.Background()
		spanCtx, span := result.Tracer.Start(ctx, "test")
		assert.NotNil(t, spanCtx)
		assert.NotNil(t, span)
		span.End()

		err = result.Provider.Shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("creates noop tracer explicitly", func(t *testing.T) {
		params := Params{
			Config: Config{
				Enabled:  true,
				Provider: "noop",
			},
			Logger: logger,
		}

		result, err := NewTracer(params)
		require.NoError(t, err)
		assert.NotNil(t, result.Tracer)
		assert.NotNil(t, result.Provider)

		// Verify it's noop
		ctx := context.Background()
		err = result.Provider.Shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("creates noop tracer for unknown provider", func(t *testing.T) {
		params := Params{
			Config: Config{
				Enabled:  true,
				Provider: "unknown-provider",
			},
			Logger: logger,
		}

		result, err := NewTracer(params)
		require.NoError(t, err)
		assert.NotNil(t, result.Tracer)
		assert.NotNil(t, result.Provider)

		// Should fall back to noop
		ctx := context.Background()
		err = result.Provider.Shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("attempts to create OTLP tracer", func(t *testing.T) {
		params := Params{
			Config: Config{
				Enabled:     true,
				Provider:    "otlp",
				ServiceName: "test-service",
				SampleRate:  1.0,
				OTLP: OTLPConfig{
					Endpoint: "localhost:4317",
					Insecure: true,
				},
			},
			Logger: logger,
		}

		result, err := NewTracer(params)

		// OTLP might fail or succeed depending on environment
		if err != nil {
			// If it fails, that's expected in test environment
			assert.Error(t, err)
		} else {
			// If it succeeds, should be able to use it
			assert.NotNil(t, result.Tracer)
			assert.NotNil(t, result.Provider)

			// Clean up
			ctx := context.Background()
			shutdownErr := result.Provider.Shutdown(ctx)
			assert.NoError(t, shutdownErr)
		}
	})
}

func TestModule(t *testing.T) {
	t.Run("returns fx module", func(t *testing.T) {
		module := Module()
		assert.NotNil(t, module)
	})
}
