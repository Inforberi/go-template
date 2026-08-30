package router

import (
	"context"
	"net/http"
	"time"

	_ "github.com/Inforberi/go-template/docs"
	"github.com/Inforberi/go-template/internal/infra/config"
	"github.com/Inforberi/go-template/internal/transport/httpx"
	appmiddleware "github.com/Inforberi/go-template/internal/transport/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

const readinessTimeout = 2 * time.Second

type HealthChecker interface {
	Ping(context.Context) error
}

type healthResponse struct {
	Status string `json:"status"`
}

func New(cfg *config.Config, log *zap.Logger, database HealthChecker) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(appmiddleware.RequestLogger(log))
	r.Use(chimiddleware.Recoverer)

	r.Get("/ping", ping)
	r.Get("/health/live", live)
	r.Get("/health/ready", ready(database))

	if cfg.Swagger.Enabled {
		swagger := r.With(chimiddleware.BasicAuth("swagger", map[string]string{
			cfg.Swagger.Username: cfg.Swagger.Password,
		}))
		swagger.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
		})
		swagger.Handle("/swagger/*", httpSwagger.Handler())
	}

	return r
}

// ping godoc
//
//	@Summary		Check API availability
//	@Description	Returns pong when the HTTP API is running.
//	@Tags			system
//	@Produce		plain
//	@Success		200	{string}	string	"pong"
//	@Router			/ping [get]
func ping(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("pong"))
}

// live godoc
//
//	@Summary	Check process liveness
//	@Tags		system
//	@Produce	json
//	@Success	200	{object}	healthResponse
//	@Router		/health/live [get]
func live(w http.ResponseWriter, _ *http.Request) {
	_ = httpx.JSONResponse(w, http.StatusOK, healthResponse{Status: "ok"})
}

// ready godoc
//
//	@Summary	Check service readiness
//	@Tags		system
//	@Produce	json
//	@Success	200	{object}	healthResponse
//	@Failure	503	{object}	healthResponse
//	@Router		/health/ready [get]
func ready(database HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()

		if err := database.Ping(ctx); err != nil {
			_ = httpx.JSONResponse(w, http.StatusServiceUnavailable, healthResponse{Status: "unavailable"})
			return
		}

		_ = httpx.JSONResponse(w, http.StatusOK, healthResponse{Status: "ok"})
	}
}
