package tracingx

import (
	"context"
	"fmt"

	"github.com/gostratum/core/logx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials/insecure"
)

// otlpProvider implements the Provider interface using OpenTelemetry
type otlpProvider struct {
	config         Config
	logger         logx.Logger
	tracer         trace.Tracer
	tracerProvider *sdktrace.TracerProvider
}

// newOTLPProvider creates a new OTLP tracing provider
func newOTLPProvider(config Config, logger logx.Logger) (Provider, error) {
	ctx := context.Background()

	// Create OTLP exporter
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.OTLP.Endpoint),
	}

	if config.OTLP.Insecure {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	if len(config.OTLP.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(config.OTLP.Headers))
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SampleRate)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for distributed tracing
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := tp.Tracer("gostratum")

	logger.Info("OTLP tracing provider initialized",
		logx.String("endpoint", config.OTLP.Endpoint),
		logx.String("service", config.ServiceName),
	)

	return &otlpProvider{
		config:         config,
		logger:         logger,
		tracer:         tracer,
		tracerProvider: tp,
	}, nil
}

// Start creates a new span
func (p *otlpProvider) Start(ctx context.Context, operationName string, opts ...SpanOption) (context.Context, Span) {
	config := applySpanOptions(opts...)

	// Convert span kind
	var otelKind trace.SpanKind
	switch config.Kind {
	case SpanKindInternal:
		otelKind = trace.SpanKindInternal
	case SpanKindServer:
		otelKind = trace.SpanKindServer
	case SpanKindClient:
		otelKind = trace.SpanKindClient
	case SpanKindProducer:
		otelKind = trace.SpanKindProducer
	case SpanKindConsumer:
		otelKind = trace.SpanKindConsumer
	default:
		otelKind = trace.SpanKindInternal
	}

	// Convert attributes
	var attrs []attribute.KeyValue
	for k, v := range config.Attributes {
		attrs = append(attrs, toAttribute(k, v))
	}

	// Start span
	spanOpts := []trace.SpanStartOption{
		trace.WithSpanKind(otelKind),
		trace.WithAttributes(attrs...),
	}

	if !config.Timestamp.IsZero() {
		spanOpts = append(spanOpts, trace.WithTimestamp(config.Timestamp))
	}

	ctx, otelSpan := p.tracer.Start(ctx, operationName, spanOpts...)

	span := &otlpSpan{
		span: otelSpan,
		ctx:  ctx,
	}

	return ContextWithSpan(ctx, span), span
}

// Extract extracts trace context from a carrier
func (p *otlpProvider) Extract(ctx context.Context, carrier any) (context.Context, error) {
	propagator := otel.GetTextMapPropagator()

	// Handle different carrier types
	var textMapCarrier propagation.TextMapCarrier
	switch c := carrier.(type) {
	case propagation.TextMapCarrier:
		textMapCarrier = c
	case map[string]string:
		textMapCarrier = propagation.MapCarrier(c)
	case map[string][]string:
		textMapCarrier = &headerCarrier{headers: c}
	default:
		return ctx, fmt.Errorf("unsupported carrier type: %T", carrier)
	}

	return propagator.Extract(ctx, textMapCarrier), nil
}

// Inject injects trace context into a carrier
func (p *otlpProvider) Inject(ctx context.Context, carrier any) error {
	propagator := otel.GetTextMapPropagator()

	// Handle different carrier types
	var textMapCarrier propagation.TextMapCarrier
	switch c := carrier.(type) {
	case propagation.TextMapCarrier:
		textMapCarrier = c
	case map[string]string:
		textMapCarrier = propagation.MapCarrier(c)
	case map[string][]string:
		textMapCarrier = &headerCarrier{headers: c}
	default:
		return fmt.Errorf("unsupported carrier type: %T", carrier)
	}

	propagator.Inject(ctx, textMapCarrier)
	return nil
}

// Shutdown shuts down the tracer provider
func (p *otlpProvider) Shutdown(ctx context.Context) error {
	if p.tracerProvider != nil {
		return p.tracerProvider.Shutdown(ctx)
	}
	return nil
}

// otlpSpan implements the Span interface
type otlpSpan struct {
	span trace.Span
	ctx  context.Context
}

func (s *otlpSpan) End() {
	s.span.End()
}

func (s *otlpSpan) SetTag(key string, value any) {
	s.span.SetAttributes(toAttribute(key, value))
}

func (s *otlpSpan) SetError(err error) {
	s.span.RecordError(err)
	s.span.SetAttributes(attribute.Bool("error", true))
}

func (s *otlpSpan) LogFields(fields ...Field) {
	attrs := make([]attribute.KeyValue, len(fields))
	for i, f := range fields {
		attrs[i] = toAttribute(f.Key, f.Value)
	}
	s.span.AddEvent("log", trace.WithAttributes(attrs...))
}

func (s *otlpSpan) Context() context.Context {
	return s.ctx
}

func (s *otlpSpan) TraceID() string {
	return s.span.SpanContext().TraceID().String()
}

func (s *otlpSpan) SpanID() string {
	return s.span.SpanContext().SpanID().String()
}

// toAttribute converts a value to an OpenTelemetry attribute
func toAttribute(key string, value any) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	case []string:
		return attribute.StringSlice(key, v)
	case []int:
		return attribute.IntSlice(key, v)
	case []int64:
		return attribute.Int64Slice(key, v)
	case []float64:
		return attribute.Float64Slice(key, v)
	case []bool:
		return attribute.BoolSlice(key, v)
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}

// headerCarrier adapts map[string][]string to propagation.TextMapCarrier
type headerCarrier struct {
	headers map[string][]string
}

func (c *headerCarrier) Get(key string) string {
	if vals, ok := c.headers[key]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (c *headerCarrier) Set(key, value string) {
	c.headers[key] = []string{value}
}

func (c *headerCarrier) Keys() []string {
	keys := make([]string, 0, len(c.headers))
	for k := range c.headers {
		keys = append(keys, k)
	}
	return keys
}
