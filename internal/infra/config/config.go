package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App
	Logger
	Swagger
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
	Username string `env:"SWAGGER_USERNAME" env-required:"true"`
	Password string `env:"SWAGGER_PASSWORD" env-required:"true"`
}

func New() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return &cfg, nil
}
