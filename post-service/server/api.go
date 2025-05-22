package server

import (
	"net/http"
	"os"

	"github.com/chrishrb/blog-microservice/internal/auth"
	"github.com/chrishrb/blog-microservice/post-service/api"
	"github.com/chrishrb/blog-microservice/post-service/config"
	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/induzo/gocom/http/middleware/writablecontext"
	"github.com/riandyrn/otelchi"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
	"k8s.io/utils/clock"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewApiHandler(settings config.ApiSettings, engine store.Engine, jwsVerifier auth.JWSVerifier) http.Handler {
	apiServer, err := api.NewServer(engine, clock.RealClock{})
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

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081"}, // Allow Swagger UI origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	if settings.Cors != nil {
		r.Use(getCorsConfig(settings.Cors).Handler)
	}

	r.Use(
		middleware.Recoverer,
		secureMiddleware.Handler,
		writablecontext.Middleware, // workaround to inject userID into chi context
		c.Handler,
		otelchi.Middleware("api", otelchi.WithChiRoutes(r)),
	)

	r.Get("/health", health)
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/post-service/openapi.json", getApiSwaggerJson)
	r.With(logger, auth.GetAuthMiddleware(swagger, jwsVerifier)).Mount("/post-service/v1", api.Handler(apiServer))
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
