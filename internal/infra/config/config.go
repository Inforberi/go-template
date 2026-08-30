package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App
	Logger
	Swagger
	Database
}

type App struct {
	App         string `env:"APP_ENV" env-required:"true"`
	ServiceName string `env:"SERVICE_NAME" env-required:"true"`
	Port        string `env:"PORT" env-required:"true"`
}

type Logger struct {
	Level string `env:"LOG_LEVEL" env-default:"info"`
}

type Swagger struct {
	Enabled  bool   `env:"SWAGGER_ENABLED" env-default:"false"`
	Username string `env:"SWAGGER_USERNAME"`
	Password string `env:"SWAGGER_PASSWORD"`
}

type Database struct {
	URL      string `env:"DATABASE_URL" env-required:"true"`
	MaxConns int32  `env:"DATABASE_MAX_CONNS" env-default:"10"`
}

func New() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

func (cfg *Config) Validate() error {
	if strings.TrimSpace(cfg.Database.URL) == "" {
		return errors.New("DATABASE_URL is required")
	}

	if cfg.Database.MaxConns < 1 {
		return errors.New("DATABASE_MAX_CONNS must be positive")
	}

	if cfg.Swagger.Enabled &&
		(strings.TrimSpace(cfg.Swagger.Username) == "" || strings.TrimSpace(cfg.Swagger.Password) == "") {
		return errors.New("SWAGGER_USERNAME and SWAGGER_PASSWORD are required when Swagger is enabled")
	}

	return nil
}
