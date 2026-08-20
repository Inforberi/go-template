package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/Inforberi/go-template/internal/infra/logger"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := chimiddleware.GetReqID(r.Context())

			requestLog := log.With(
				zap.String("request_id", requestID),
			)

			ctx := logger.WithContext(r.Context(), requestLog)
			r = r.WithContext(ctx)

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			fields := []zap.Field{
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", status),
				zap.Int("bytes", ww.BytesWritten()),
				zap.Int64("duration_ms", time.Since(start).Milliseconds()),
			}

			switch {
			case status >= http.StatusInternalServerError:
				log.Error("request completed", fields...)
			case status >= http.StatusBadRequest:
				log.Warn("request completed", fields...)
			default:
				log.Info("request completed", fields...)
			}
		})
	}
}
