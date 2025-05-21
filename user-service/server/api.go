package server

import (
	"net/http"
	"os"

	"github.com/chrishrb/blog-microservice/internal/auth"
	"github.com/chrishrb/blog-microservice/user-service/api"
	"github.com/chrishrb/blog-microservice/user-service/config"
	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/chrishrb/blog-microservice/internal/writeablecontext"
	"github.com/riandyrn/otelchi"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
	"k8s.io/utils/clock"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewApiHandler(settings config.ApiSettings, engine store.Engine, JWSVerifier auth.JWSVerifier, JWSSigner auth.JWSSigner) http.Handler {
	apiServer, err := api.NewServer(engine, clock.RealClock{}, JWSVerifier, JWSSigner)
	if err != nil {
		panic(err)
	}

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

	logger := middleware.RequestLogger(logFormatter{endpoint: "api"})
	swagger, _ := api.GetSwagger()

	if settings.Cors != nil {
		r.Use(getCorsConfig(settings.Cors).Handler)
	}

	r.Use(
		middleware.Recoverer,
		secureMiddleware.Handler,
		writeablecontext.Middleware, // workaround to inject userID into chi context
		otelchi.Middleware("api", otelchi.WithChiRoutes(r)),
	)

	r.Get("/health", health)
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/user-service/openapi.json", getApiSwaggerJson)
	r.With(logger, auth.GetAuthMiddleware(swagger, JWSVerifier)).Mount("/user-service/v1", api.Handler(apiServer))
	return r
}

func getApiSwaggerJson(w http.ResponseWriter, r *http.Request) {
	swagger, err := api.GetSwagger()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	json, err := swagger.MarshalJSON()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(json)
}

func getCorsConfig(cfg *config.CorsConfig) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		AllowCredentials: true,
	})
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"OK"}`))
}
