package tracingx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigStructure(t *testing.T) {
	t.Run("config has correct prefix", func(t *testing.T) {
		cfg := Config{}
		assert.Equal(t, "tracing", cfg.Prefix())
	})

	t.Run("config fields validation", func(t *testing.T) {
		cfg := Config{
			Enabled:     true,
			ServiceName: "test-service",
			Provider:    "otlp",
			SampleRate:  0.5,
			OTLP: OTLPConfig{
				Endpoint: "localhost:4317",
				Insecure: true,
				Headers: map[string]string{
					"api-key": "secret",
				},
			},
			Jaeger: JaegerConfig{
				Endpoint:  "http://localhost:14268/api/traces",
				AgentHost: "localhost",
				AgentPort: "6831",
			},
		}

		assert.True(t, cfg.Enabled)
		assert.Equal(t, "test-service", cfg.ServiceName)
		assert.Equal(t, "otlp", cfg.Provider)
		assert.Equal(t, 0.5, cfg.SampleRate)
		assert.Equal(t, "localhost:4317", cfg.OTLP.Endpoint)
		assert.True(t, cfg.OTLP.Insecure)
		assert.Equal(t, "secret", cfg.OTLP.Headers["api-key"])
		assert.Equal(t, "http://localhost:14268/api/traces", cfg.Jaeger.Endpoint)
		assert.Equal(t, "localhost", cfg.Jaeger.AgentHost)
		assert.Equal(t, "6831", cfg.Jaeger.AgentPort)
	})
}

func TestOTLPConfig(t *testing.T) {
	t.Run("OTLP config with custom headers", func(t *testing.T) {
		cfg := OTLPConfig{
			Endpoint: "otel-collector:4317",
			Insecure: false,
			Headers: map[string]string{
				"authorization": "Bearer token123",
				"tenant-id":     "tenant-abc",
			},
		}

		assert.Equal(t, "otel-collector:4317", cfg.Endpoint)
		assert.False(t, cfg.Insecure)
		assert.Len(t, cfg.Headers, 2)
		assert.Equal(t, "Bearer token123", cfg.Headers["authorization"])
		assert.Equal(t, "tenant-abc", cfg.Headers["tenant-id"])
	})

	t.Run("OTLP config without headers", func(t *testing.T) {
		cfg := OTLPConfig{
			Endpoint: "localhost:4317",
			Insecure: true,
		}

		assert.Equal(t, "localhost:4317", cfg.Endpoint)
		assert.True(t, cfg.Insecure)
		assert.Nil(t, cfg.Headers)
	})
}

func TestJaegerConfig(t *testing.T) {
	t.Run("Jaeger config with all fields", func(t *testing.T) {
		cfg := JaegerConfig{
			Endpoint:  "http://jaeger-collector:14268/api/traces",
			AgentHost: "jaeger-agent",
			AgentPort: "6831",
		}

		assert.Equal(t, "http://jaeger-collector:14268/api/traces", cfg.Endpoint)
		assert.Equal(t, "jaeger-agent", cfg.AgentHost)
		assert.Equal(t, "6831", cfg.AgentPort)
	})
}

func TestSampleRate(t *testing.T) {
	t.Run("sample rate validation", func(t *testing.T) {
		testCases := []struct {
			name       string
			sampleRate float64
			valid      bool
		}{
			{"zero sample rate", 0.0, true},
			{"half sample rate", 0.5, true},
			{"full sample rate", 1.0, true},
			{"negative sample rate", -0.1, false},
			{"over one sample rate", 1.5, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := Config{SampleRate: tc.sampleRate}

				// Valid sample rates should be between 0.0 and 1.0
				if tc.valid {
					assert.GreaterOrEqual(t, cfg.SampleRate, 0.0)
					assert.LessOrEqual(t, cfg.SampleRate, 1.0)
				} else {
					// Invalid rates should be outside this range
					assert.True(t, cfg.SampleRate < 0.0 || cfg.SampleRate > 1.0)
				}
			})
		}
	})
}
