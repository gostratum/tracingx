# tracingx

**Distributed tracing module for gostratum framework**

`tracingx` provides distributed tracing capabilities using OpenTelemetry, enabling you to trace requests across services and understand your application's behavior in production. It follows the gostratum philosophy of fx-first design and integrates seamlessly with other modules.

## Features

- 🔍 **OpenTelemetry-based** - Industry-standard distributed tracing
- 🎯 **Provider-agnostic** - Support for OTLP, Jaeger, and custom providers
- 🔧 **Fx-first design** - Seamless dependency injection
- 📝 **Contextual logging** - Correlate traces with logs
- 🌐 **Distributed context** - Automatic propagation across services
- 🎨 **Clean API** - Simple and intuitive span management
- 🧪 **Testable** - Includes no-op provider for testing

## Installation

```bash
go get github.com/gostratum/tracingx
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "go.uber.org/fx"
    
    "github.com/gostratum/core"
    "github.com/gostratum/tracingx"
)

func main() {
    fx.New(
        core.Module(),
        tracingx.Module(),
        
        fx.Invoke(func(tracer tracingx.Tracer) {
            ctx := context.Background()
            
            // Start a span
            ctx, span := tracer.Start(ctx, "operation",
                tracingx.WithSpanKind(tracingx.SpanKindServer),
                tracingx.WithAttributes(map[string]interface{}{
                    "user.id": "12345",
                    "action":  "create_order",
                }),
            )
            defer span.End()
            
            // Add tags
            span.SetTag("order.id", "ord-123")
            
            // Log events
            span.LogFields(
                tracingx.Field{Key: "event", Value: "validation_complete"},
                tracingx.Field{Key: "duration_ms", Value: 42},
            )
            
            // Record errors
            if err := doSomething(); err != nil {
                span.SetError(err)
            }
        }),
    ).Run()
}
```

### Configuration

Add to your `config.yaml`:

```yaml
tracing:
  enabled: true
  service_name: my-service
  provider: otlp
  sample_rate: 1.0  # 100% sampling (adjust for production)
  
  otlp:
    endpoint: localhost:4317
    insecure: true
    headers:
      api-key: your-api-key
```

## Span Types

### Server Span (incoming request)

```go
ctx, span := tracer.Start(ctx, "HandleRequest",
    tracingx.WithSpanKind(tracingx.SpanKindServer),
    tracingx.WithAttributes(map[string]interface{}{
        "http.method": "POST",
        "http.url":    "/api/orders",
    }),
)
defer span.End()
```

### Client Span (outgoing request)

```go
ctx, span := tracer.Start(ctx, "CallExternalAPI",
    tracingx.WithSpanKind(tracingx.SpanKindClient),
    tracingx.WithAttributes(map[string]interface{}{
        "http.url":    "https://api.example.com/data",
        "http.method": "GET",
    }),
)
defer span.End()
```

### Internal Span (internal operation)

```go
ctx, span := tracer.Start(ctx, "ProcessData",
    tracingx.WithSpanKind(tracingx.SpanKindInternal),
)
defer span.End()
```

### Producer/Consumer Spans (messaging)

```go
// Producer
ctx, span := tracer.Start(ctx, "PublishMessage",
    tracingx.WithSpanKind(tracingx.SpanKindProducer),
    tracingx.WithAttributes(map[string]interface{}{
        "messaging.system":      "kafka",
        "messaging.destination": "orders.created",
    }),
)
defer span.End()

// Consumer
ctx, span := tracer.Start(ctx, "ConsumeMessage",
    tracingx.WithSpanKind(tracingx.SpanKindConsumer),
)
defer span.End()
```

## Distributed Tracing

### HTTP Server (Extract incoming trace)

```go
// In your HTTP handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Extract trace context from headers
    ctx, err := h.tracer.Extract(ctx, r.Header)
    if err != nil {
        // Continue without parent trace
    }
    
    // Start server span
    ctx, span := h.tracer.Start(ctx, "HandleHTTP",
        tracingx.WithSpanKind(tracingx.SpanKindServer),
    )
    defer span.End()
    
    // Use ctx for downstream operations
    h.processRequest(ctx, r)
}
```

### HTTP Client (Inject outgoing trace)

```go
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
    // Start client span
    ctx, span := c.tracer.Start(ctx, "HTTPRequest",
        tracingx.WithSpanKind(tracingx.SpanKindClient),
    )
    defer span.End()
    
    // Inject trace context into headers
    if err := c.tracer.Inject(ctx, req.Header); err != nil {
        span.SetError(err)
    }
    
    // Make request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        span.SetError(err)
        return nil, err
    }
    
    span.SetTag("http.status_code", resp.StatusCode)
    return resp, nil
}
```

## Integration with httpx

Automatic HTTP tracing middleware:

```go
package main

import (
    "go.uber.org/fx"
    
    "github.com/gostratum/core"
    "github.com/gostratum/httpx"
    "github.com/gostratum/tracingx"
)

func main() {
    fx.New(
        core.Module(),
        tracingx.Module(),
        httpx.Module(),
        // httpx automatically detects tracingx and adds middleware
    ).Run()
}
```

This automatically:
- Extracts trace context from incoming requests
- Creates server spans for each request
- Propagates context to handlers
- Records request details and errors

## Log Correlation

Enrich logs with trace information:

```go
import (
    "github.com/gostratum/core/logx"
    "github.com/gostratum/tracingx"
)

func HandleRequest(ctx context.Context, logger logx.Logger) {
    // Extract span from context
    span := tracingx.SpanFromContext(ctx)
    if span != nil {
        // Add trace IDs to logger
        logger = logger.With(
            logx.String("trace_id", span.TraceID()),
            logx.String("span_id", span.SpanID()),
        )
    }
    
    logger.Info("processing request")
    // Log output: {"msg":"processing request","trace_id":"abc123","span_id":"xyz789"}
}
```

## Custom Attributes

### Standard Attributes

```go
span.SetTag("http.method", "POST")
span.SetTag("http.status_code", 200)
span.SetTag("http.url", "/api/orders")
span.SetTag("db.system", "postgresql")
span.SetTag("db.statement", "SELECT * FROM orders")
span.SetTag("user.id", "12345")
```

### Business Attributes

```go
span.SetTag("order.id", "ord-12345")
span.SetTag("order.total", 99.99)
span.SetTag("customer.tier", "premium")
span.SetTag("feature.flag", true)
```

## Error Tracking

```go
ctx, span := tracer.Start(ctx, "ProcessOrder")
defer span.End()

if err := processOrder(ctx, order); err != nil {
    // Record the error
    span.SetError(err)
    
    // Add additional context
    span.SetTag("error.type", "validation_error")
    span.LogFields(
        tracingx.Field{Key: "error.message", Value: err.Error()},
        tracingx.Field{Key: "order.id", Value: order.ID},
    )
    
    return err
}
```

## Sampling

Control sampling rate to reduce overhead:

```yaml
tracing:
  sample_rate: 0.1  # Sample 10% of traces
```

Or use adaptive sampling based on conditions:

```go
// Always trace errors
if err != nil {
    span.SetTag("sampled", true)
}

// Always trace slow requests
if duration > time.Second {
    span.SetTag("sampled", true)
}
```

## Providers

### OTLP (Default)

OpenTelemetry Protocol - works with Jaeger, Tempo, etc:

```yaml
tracing:
  provider: otlp
  otlp:
    endpoint: localhost:4317
    insecure: true
```

Start OTLP collector:
```bash
docker run -p 4317:4317 otel/opentelemetry-collector
```

### Jaeger

Direct Jaeger integration:

```yaml
tracing:
  provider: jaeger
  jaeger:
    endpoint: http://localhost:14268/api/traces
    agent_host: localhost
    agent_port: 6831
```

Start Jaeger:
```bash
docker run -p 16686:16686 -p 14268:14268 -p 6831:6831/udp jaegertracing/all-in-one
```

Access UI: http://localhost:16686

### No-op Provider

For testing and development:

```yaml
tracing:
  enabled: false
```

Or:

```yaml
tracing:
  provider: noop
```

## Best Practices

### 1. **Span Naming**

Use consistent, descriptive names:

```go
// ✅ GOOD
tracer.Start(ctx, "GetUserByID")
tracer.Start(ctx, "CreateOrder")
tracer.Start(ctx, "HTTP POST /api/orders")

// ❌ BAD
tracer.Start(ctx, "operation")
tracer.Start(ctx, "func1")
```

### 2. **Context Propagation**

Always pass context through your call chain:

```go
// ✅ GOOD
func HandleRequest(ctx context.Context) error {
    ctx, span := tracer.Start(ctx, "HandleRequest")
    defer span.End()
    
    return processData(ctx)  // Pass ctx
}

// ❌ BAD
func HandleRequest(ctx context.Context) error {
    ctx, span := tracer.Start(ctx, "HandleRequest")
    defer span.End()
    
    return processData(context.Background())  // Lost context!
}
```

### 3. **Defer span.End()**

Always defer span.End() immediately after creation:

```go
// ✅ GOOD
ctx, span := tracer.Start(ctx, "Operation")
defer span.End()
// ... rest of code

// ❌ BAD
ctx, span := tracer.Start(ctx, "Operation")
// ... code ...
span.End()  // Might not be called if error occurs
```

### 4. **Add Relevant Attributes**

Add attributes that help debugging:

```go
span.SetTag("user.id", userID)
span.SetTag("order.id", orderID)
span.SetTag("payment.method", "credit_card")
span.SetTag("cache.hit", true)
```

### 5. **Record Errors**

Always record errors in spans:

```go
if err != nil {
    span.SetError(err)
    span.SetTag("error.message", err.Error())
}
```

## Testing

Use the no-op provider in tests:

```go
func TestOrderService(t *testing.T) {
    tracer := tracingx.NewNoopProvider()
    
    service := NewOrderService(tracer)
    // ... test your service ...
}
```

Or inject a test tracer:

```go
import "github.com/gostratum/tracingx"

func NewTestTracer() tracingx.Tracer {
    config := tracingx.Config{
        Enabled:  false,
        Provider: "noop",
    }
    // Create with noop provider
}
```

## Performance

### Overhead

- No-op tracer: ~5ns per operation (essentially zero)
- OTLP tracer: ~500ns-1µs per span
- Sampling reduces overhead proportionally

### Production Tips

```yaml
tracing:
  # Sample only 10% in high-traffic services
  sample_rate: 0.1
  
  # Use batch exporter (already configured)
  otlp:
    endpoint: collector:4317
```

## Architecture

```
tracingx/
├── tracer.go            # Core interfaces
├── config.go            # Configuration
├── module.go            # Fx module
├── provider_otlp.go     # OpenTelemetry implementation
└── provider_noop.go     # No-op implementation
```

## Dependencies

- **Core**: `github.com/gostratum/core` (for config and logging)
- **OpenTelemetry**: `go.opentelemetry.io/otel` (tracing standard)
- **Fx**: `go.uber.org/fx` (dependency injection)

## Roadmap

- [ ] Jaeger native exporter
- [ ] Zipkin exporter
- [ ] Span baggage support
- [ ] Trace sampling strategies
- [ ] gRPC middleware
- [ ] Database instrumentation helpers

## Examples

See [examples/](./examples/) for complete examples:
- Basic tracing
- HTTP service with distributed tracing
- Microservice communication
- Log correlation

## License

MIT License - see LICENSE file for details

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## Related Modules

- [`core`](../core/README.md) - Foundation module
- [`metricsx`](../metricsx/README.md) - Metrics and monitoring
- [`httpx`](../httpx/README.md) - HTTP server with automatic tracing
- [`dbx`](../dbx/README.md) - Database operations
