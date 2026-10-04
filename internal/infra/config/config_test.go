package config

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	validConfig := func() Config {
		return Config{
			App:      App{Port: "8080"},
			Database: Database{URL: "postgres://app:secret@postgres:5432/app", MaxConns: 10},
		}
	}

	tests := []struct {
		name      string
		config    Config
		wantError string
	}{
		{
			name:   "valid",
			config: validConfig(),
		},
		{
			name: "empty port",
			config: func() Config {
				cfg := validConfig()
				cfg.Port = ""
				return cfg
			}(),
			wantError: "PORT",
		},
		{
			name: "non-numeric port",
			config: func() Config {
				cfg := validConfig()
				cfg.Port = "http"
				return cfg
			}(),
			wantError: "PORT",
		},
		{
			name: "zero port",
			config: func() Config {
				cfg := validConfig()
				cfg.Port = "0"
				return cfg
			}(),
			wantError: "PORT",
		},
		{
			name: "negative port",
			config: func() Config {
				cfg := validConfig()
				cfg.Port = "-1"
				return cfg
			}(),
			wantError: "PORT",
		},
		{
			name: "port above maximum",
			config: func() Config {
				cfg := validConfig()
				cfg.Port = "65536"
				return cfg
			}(),
			wantError: "PORT",
		},
		{
			name:      "database URL",
			config:    Config{App: App{Port: "8080"}, Database: Database{MaxConns: 10}},
			wantError: "DATABASE_URL",
		},
		{
			name:      "maximum connections",
			config:    Config{App: App{Port: "8080"}, Database: Database{URL: "postgres://postgres", MaxConns: 0}},
			wantError: "DATABASE_MAX_CONNS",
		},
		{
			name: "Swagger credentials",
			config: Config{
				App:      App{Port: "8080"},
				Database: Database{URL: "postgres://postgres", MaxConns: 10},
				Swagger:  Swagger{Enabled: true},
			},
			wantError: "SWAGGER_USERNAME",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if test.wantError == "" && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if test.wantError != "" && (err == nil || !strings.Contains(err.Error(), test.wantError)) {
				t.Fatalf("Validate() error = %v, want substring %q", err, test.wantError)
			}
		})
	}
}
