package router

import (
	"net/http"

	_ "github.com/Inforberi/go-template/docs"
	"github.com/Inforberi/go-template/internal/infra/config"
	appmiddleware "github.com/Inforberi/go-template/internal/transport/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

func New(cfg *config.Config, log *zap.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(appmiddleware.RequestLogger(log))
	r.Use(chimiddleware.Recoverer)

	r.Get("/ping", ping)
	swagger := r.With(chimiddleware.BasicAuth("swagger", map[string]string{
		cfg.Swagger.Username: cfg.Swagger.Password,
	}))
	swagger.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
	})
	swagger.Handle("/swagger/*", httpSwagger.Handler())

	return r
}

// ping godoc
//
//	@Summary		Check API availability
//	@Description	Returns pong when the HTTP API is running.
//	@Tags			system
//	@Produce		plain
//	@Success		200	{string}	string	"pongg"
//	@Router			/ping [get]
func ping(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("pongg"))
}
