package tracingx

import (
	"github.com/gostratum/core/configx"
)

// Config contains configuration for the tracing module
type Config struct {
	// Enabled determines if tracing is enabled
	Enabled bool `mapstructure:"enabled" default:"true"`

	// ServiceName identifies this service in traces
	ServiceName string `mapstructure:"service_name" default:"gostratum-service"`

	// Provider specifies which tracing provider to use (otlp, jaeger, noop)
	Provider string `mapstructure:"provider" default:"otlp"`

	// SampleRate determines the sampling rate (0.0 to 1.0)
	SampleRate float64 `mapstructure:"sample_rate" default:"1.0"`

	// OTLP configuration
	OTLP OTLPConfig `mapstructure:"otlp"`

	// Jaeger configuration
	Jaeger JaegerConfig `mapstructure:"jaeger"`
}

// Prefix enables configx.Bind
func (Config) Prefix() string { return "tracing" }

// OTLPConfig contains OpenTelemetry Protocol configuration
type OTLPConfig struct {
	// Endpoint is the OTLP receiver endpoint
	Endpoint string `mapstructure:"endpoint" default:"localhost:4317"`

	// Insecure determines if the connection should be insecure
	Insecure bool `mapstructure:"insecure" default:"true"`

	// Headers are additional headers to send with requests
	Headers map[string]string `mapstructure:"headers"`
}

// JaegerConfig contains Jaeger-specific configuration
type JaegerConfig struct {
	// Endpoint is the Jaeger collector endpoint
	Endpoint string `mapstructure:"endpoint" default:"http://localhost:14268/api/traces"`

	// AgentHost is the Jaeger agent host
	AgentHost string `mapstructure:"agent_host" default:"localhost"`

	// AgentPort is the Jaeger agent port
	AgentPort string `mapstructure:"agent_port" default:"6831"`
}

// NewConfig creates a new Config from the configuration loader
func NewConfig(loader configx.Loader) (Config, error) {
	var cfg Config
	if err := loader.Bind(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
