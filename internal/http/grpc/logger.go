package grpc

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
)

func LoggingInterceptor(ctx context.Context, log *slog.Logger, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	log = log.With(slog.String("component", "middleware/logger"))
	
	log.Info("Method logging started", "info", info.FullMethod)

	resp, err := handler(ctx, req)

	log.Info("Method ended", "info", info.FullMethod, "time", time.Since(start), "error", err)

	return resp, err
}

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(slog.String("component", "middleware/logger"))

		log.Info("Logger middleware initialized")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote", r.RemoteAddr),
				slog.String("user-agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				duration := time.Since(t1)

				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bites", ww.BytesWritten()),
					slog.String("duration", duration.String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
