package subscription

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	logmid "test_task/internal/middleware/loger_middleware"
	"test_task/swagger"
)

func Router(log *slog.Logger, h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(logmid.RequestLogger(log))

	r.Get("/healthz", h.Healthz)

	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(swagger.SwaggerJSON)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/subscriptions", func(r chi.Router) {
			r.Get("/", h.ListSubscriptions)
			r.Post("/", h.CreateSubscription)

			r.Get("/summary", h.Summary)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetSubscription)
				r.Put("/", h.UpdateSubscription)
				r.Delete("/", h.DeleteSubscription)
			})
		})
	})

	return r
}
