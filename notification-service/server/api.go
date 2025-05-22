package server

import (
	"net/http"
	"os"

	"github.com/riandyrn/otelchi"
	"github.com/rs/cors"
	"github.com/unrolled/secure"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewApiHandler() http.Handler {
	var isDevelopment bool
	if os.Getenv("ENVIRONMENT") == "dev" {
		isDevelopment = true
	}

	secureMiddleware := secure.New(secure.Options{
		IsDevelopment:         isDevelopment,
		BrowserXssFilter:      true,
		ContentTypeNosniff:    true,
		FrameDeny:             true,
		ContentSecurityPolicy: "frame-ancestors: 'none'",
	})

	r := chi.NewRouter()

	r.Use(
		middleware.Recoverer,
		secureMiddleware.Handler,
		cors.Default().Handler,
		otelchi.Middleware("api", otelchi.WithChiRoutes(r)),
	)

	r.Get("/health", health)
	r.Handle("/metrics", promhttp.Handler())
	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"OK"}`))
}
