package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Inforberi/go-template/internal/infra/config"
	"go.uber.org/zap"
)

type fakeDatabase struct {
	err error
}

func (database fakeDatabase) Ping(context.Context) error {
	return database.err
}

func testConfig(swaggerEnabled bool) *config.Config {
	return &config.Config{
		Swagger: config.Swagger{
			Enabled:  swaggerEnabled,
			Username: "swagger",
			Password: "secret",
		},
	}
}

func TestSystemEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		database   fakeDatabase
		wantStatus int
		wantBody   string
	}{
		{name: "ping", path: "/ping", wantStatus: http.StatusOK, wantBody: "pong"},
		{name: "live", path: "/health/live", wantStatus: http.StatusOK, wantBody: `{"status":"ok"}` + "\n"},
		{name: "ready", path: "/health/ready", wantStatus: http.StatusOK, wantBody: `{"status":"ok"}` + "\n"},
		{
			name:       "not ready",
			path:       "/health/ready",
			database:   fakeDatabase{err: errors.New("unavailable")},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `{"status":"unavailable"}` + "\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := New(testConfig(false), zap.NewNop(), test.database)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))

			if recorder.Code != test.wantStatus || recorder.Body.String() != test.wantBody {
				t.Fatalf("response = (%d, %q), want (%d, %q)", recorder.Code, recorder.Body.String(), test.wantStatus, test.wantBody)
			}
		})
	}
}

func TestSwaggerToggleAndAuthentication(t *testing.T) {
	disabled := New(testConfig(false), zap.NewNop(), fakeDatabase{})
	recorder := httptest.NewRecorder()
	disabled.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/swagger", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("disabled Swagger status = %d", recorder.Code)
	}

	enabled := New(testConfig(true), zap.NewNop(), fakeDatabase{})
	recorder = httptest.NewRecorder()
	enabled.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/swagger", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated Swagger status = %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	request.SetBasicAuth("swagger", "secret")
	enabled.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusFound {
		t.Fatalf("authenticated Swagger status = %d", recorder.Code)
	}
}
