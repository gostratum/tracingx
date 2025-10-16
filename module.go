package tracingx

import (
	"context"

	"github.com/gostratum/core/logx"
	"go.uber.org/fx"
)

// Params contains dependencies for the tracing module
type Params struct {
	fx.In
	Config Config
	Logger logx.Logger
}

// Result contains outputs from the tracing module
type Result struct {
	fx.Out
	Tracer   Tracer
	Provider Provider
}

// Module provides the tracing module for fx
func Module() fx.Option {
	return fx.Module("tracingx",
		fx.Provide(
			NewConfig,
			NewTracer,
		),
		fx.Invoke(registerLifecycle),
	)
}

// NewTracer creates a new Tracer instance based on configuration
func NewTracer(p Params) (Result, error) {
	if !p.Config.Enabled {
		p.Logger.Info("tracing is disabled, using noop tracer")
		provider := newNoopProvider()
		return Result{
			Tracer:   provider,
			Provider: provider,
		}, nil
	}

	var provider Provider
	var err error

	switch p.Config.Provider {
	case "otlp":
		provider, err = newOTLPProvider(p.Config, p.Logger)
	case "noop":
		provider = newNoopProvider()
	default:
		p.Logger.Warn("unknown tracing provider, using noop", logx.String("provider", p.Config.Provider))
		provider = newNoopProvider()
	}

	if err != nil {
		return Result{}, err
	}

	return Result{
		Tracer:   provider,
		Provider: provider,
	}, nil
}

// registerLifecycle registers the tracing lifecycle hooks
func registerLifecycle(lc fx.Lifecycle, provider Provider, logger logx.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting tracing provider")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping tracing provider")
			return provider.Shutdown(ctx)
		},
	})
}
