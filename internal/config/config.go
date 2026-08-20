package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App
	Logger
}

type App struct {
	App         string `env:"APP_ENV" env-required:"true"`
	ServiceName string `env:"SERVICE_NAME" env-required:"true"`
	Port        string `env:"PORT" env-required:"true"`
}

type Logger struct {
	Level string `env:"LOG_LEVEL" default:"info"`
}

func New() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		return nil, fmt.Errorf("fail to read config: %w", err)
	}

	return &cfg, nil
}
