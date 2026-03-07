package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *AnalyticsHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Use(LoggingMiddleware)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/projects/{projectID}", func(r chi.Router) {
			r.Route("/analytics", func(r chi.Router) {
				r.Get("/", h.GetAnalytics)
				r.Post("/recalculate", h.Recalculate)
			})
		})
	})
	return r
}
