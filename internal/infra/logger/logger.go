package logger

import (
	"fmt"
	"strings"
	"time"

	"github.com/Inforberi/go-template/internal/infra/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg *config.Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(strings.ToLower(strings.TrimSpace(cfg.Level)))
	if err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", cfg.Level, err)
	}

	var zapCfg zap.Config

	switch cfg.App.App {
	case "dev":
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("02-01-2006 15:04:05"))
		}

	case "prod":
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	default:
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	zapCfg.Level = zap.NewAtomicLevelAt(level)

	zapCfg.EncoderConfig.TimeKey = "timestamp"

	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}

	zapCfg.InitialFields = map[string]any{
		"service": cfg.ServiceName,
		"env":     cfg.App.App,
	}

	log, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return log, nil
}
