package tracingx

import (
	"context"
)

// noopProvider implements a no-op tracing provider for testing
type noopProvider struct{}

// newNoopProvider creates a new no-op provider
func newNoopProvider() Provider {
	return &noopProvider{}
}

func (p *noopProvider) Start(ctx context.Context, operationName string, opts ...SpanOption) (context.Context, Span) {
	span := &noopSpan{ctx: ctx}
	return ContextWithSpan(ctx, span), span
}

func (p *noopProvider) Extract(ctx context.Context, carrier any) (context.Context, error) {
	return ctx, nil
}

func (p *noopProvider) Inject(ctx context.Context, carrier any) error {
	return nil
}

func (p *noopProvider) Shutdown(ctx context.Context) error {
	return nil
}

// noopSpan implements the Span interface
type noopSpan struct {
	ctx context.Context
}

func (s *noopSpan) End()                         {}
func (s *noopSpan) SetTag(key string, value any) {}
func (s *noopSpan) SetError(err error)           {}
func (s *noopSpan) LogFields(fields ...Field)    {}
func (s *noopSpan) Context() context.Context     { return s.ctx }
func (s *noopSpan) TraceID() string              { return "" }
func (s *noopSpan) SpanID() string               { return "" }
