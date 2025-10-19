package logger

import (
	"context"
	"strings"

	"github.com/traweezy/tacticboard/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module exposes the Fx module wiring a zap.Logger with lifecycle managed sync.
var Module = fx.Module(
	"logger",
	fx.Provide(New),
	fx.Invoke(registerLifecycle),
)

// New constructs a zap logger tuned for the provided environment.
func New(cfg config.Config) (*zap.Logger, error) {
	var (
		log *zap.Logger
		err error
	)

	switch strings.ToLower(cfg.Environment) {
	case "development", "dev", "local":
		log, err = zap.NewDevelopment()
	default:
		zapCfg := zap.NewProductionConfig()
		zapCfg.EncoderConfig.TimeKey = "timestamp"
		zapCfg.EncoderConfig.MessageKey = "message"
		zapCfg.DisableCaller = false
		log, err = zapCfg.Build()
	}

	if err != nil {
		return nil, err
	}

	return log.Named("tacticboard").With(zap.String("env", cfg.Environment)), nil
}

func registerLifecycle(lc fx.Lifecycle, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return log.Sync()
		},
	})
}
