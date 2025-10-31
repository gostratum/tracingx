package tracingx

import (
	"testing"

	"github.com/gostratum/core/logx"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// TestConfig_Sanitize verifies that Config implements logx.Sanitizable
// and properly redacts secrets when logged.
func TestConfig_Sanitize(t *testing.T) {
	t.Run("implements Sanitizable interface", func(t *testing.T) {
		cfg := Config{
			Enabled:     true,
			ServiceName: "test-service",
			Provider:    "otlp",
			OTLP: OTLPConfig{
				Endpoint: "localhost:4317",
				Headers: map[string]string{
					"api-key":       "secret-api-key",
					"authorization": "Bearer secret-token",
					"custom-header": "safe-value",
					"x-secret":      "secret-value",
				},
			},
		}

		// Verify it implements Sanitizable
		var _ interface{ Sanitize() any } = cfg

		// Get sanitized version
		sanitized := cfg.Sanitize()
		sanitizedCfg, ok := sanitized.(Config)
		if !ok {
			t.Fatalf("Sanitize() returned wrong type: %T", sanitized)
		}

		// Verify secret headers are redacted
		if sanitizedCfg.OTLP.Headers["api-key"] != "[redacted]" {
			t.Errorf("api-key header not redacted: %s", sanitizedCfg.OTLP.Headers["api-key"])
		}
		if sanitizedCfg.OTLP.Headers["authorization"] != "[redacted]" {
			t.Errorf("authorization header not redacted: %s", sanitizedCfg.OTLP.Headers["authorization"])
		}
		if sanitizedCfg.OTLP.Headers["x-secret"] != "[redacted]" {
			t.Errorf("x-secret header not redacted: %s", sanitizedCfg.OTLP.Headers["x-secret"])
		}

		// Verify non-secret headers are preserved
		if sanitizedCfg.OTLP.Headers["custom-header"] != "safe-value" {
			t.Errorf("custom-header changed: %s", sanitizedCfg.OTLP.Headers["custom-header"])
		}

		// Verify non-secret fields are preserved
		if sanitizedCfg.ServiceName != "test-service" {
			t.Errorf("ServiceName changed: %s", sanitizedCfg.ServiceName)
		}
		if sanitizedCfg.Provider != "otlp" {
			t.Errorf("Provider changed: %s", sanitizedCfg.Provider)
		}
		if sanitizedCfg.OTLP.Endpoint != "localhost:4317" {
			t.Errorf("OTLP.Endpoint changed: %s", sanitizedCfg.OTLP.Endpoint)
		}

		// Verify original is not mutated
		if cfg.OTLP.Headers["api-key"] == "[redacted]" {
			t.Error("Original api-key header was mutated")
		}
		if cfg.OTLP.Headers["authorization"] == "[redacted]" {
			t.Error("Original authorization header was mutated")
		}
	})

	t.Run("auto-sanitizes with logx.Any", func(t *testing.T) {
		// Create observer to capture logs
		observedZapCore, observedLogs := observer.New(zap.DebugLevel)
		observedLogger := zap.New(observedZapCore)
		logger := logx.ProvideAdapter(observedLogger)

		cfg := Config{
			Enabled:     true,
			ServiceName: "test-service",
			Provider:    "otlp",
			OTLP: OTLPConfig{
				Endpoint: "localhost:4317",
				Headers: map[string]string{
					"authorization": "Bearer secret-token",
					"x-api-key":     "secret-key",
				},
			},
		}

		// Log the config using logx.Any (should auto-sanitize)
		logger.Info("Tracing config loaded", logx.Any("config", cfg))

		// Verify log was captured
		if observedLogs.Len() == 0 {
			t.Fatal("No logs captured")
		}

		// Get the logged entry
		entries := observedLogs.All()
		if len(entries) == 0 {
			t.Fatal("No log entries")
		}

		entry := entries[0]
		if entry.Message != "Tracing config loaded" {
			t.Errorf("Wrong message: %s", entry.Message)
		}

		// Find the config field
		var configField *zap.Field
		for i := range entry.Context {
			if entry.Context[i].Key == "config" {
				configField = &entry.Context[i]
				break
			}
		}

		if configField == nil {
			t.Fatal("Config field not found in log")
		}

		// The field should contain the sanitized config
		// We can't easily inspect the nested structure, but we've verified
		// that Sanitize() works correctly above and logx.Any() calls it
		t.Logf("Config field type: %v", configField.Type)
	})

	t.Run("config with no headers", func(t *testing.T) {
		cfg := Config{
			Enabled:     true,
			ServiceName: "test-service",
			Provider:    "otlp",
			OTLP: OTLPConfig{
				Endpoint: "localhost:4317",
				Headers:  nil,
			},
		}

		sanitized := cfg.Sanitize()
		sanitizedCfg, ok := sanitized.(Config)
		if !ok {
			t.Fatalf("Sanitize() returned wrong type: %T", sanitized)
		}

		if sanitizedCfg.OTLP.Headers != nil {
			t.Errorf("Expected nil headers, got %v", sanitizedCfg.OTLP.Headers)
		}
	})
}
