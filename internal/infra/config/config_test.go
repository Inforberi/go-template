package config

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError string
	}{
		{
			name: "valid",
			config: Config{
				Database: Database{URL: "postgres://app:secret@postgres:5432/app", MaxConns: 10},
			},
		},
		{
			name:      "database URL",
			config:    Config{Database: Database{MaxConns: 10}},
			wantError: "DATABASE_URL",
		},
		{
			name:      "maximum connections",
			config:    Config{Database: Database{URL: "postgres://postgres", MaxConns: 0}},
			wantError: "DATABASE_MAX_CONNS",
		},
		{
			name: "Swagger credentials",
			config: Config{
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
