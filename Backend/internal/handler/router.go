package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/microservices-development-hse/backend/internal/config"
	"github.com/rs/cors"
)

func NewRouter(cfg *config.Config, projectHandler *ProjectHandler, analyticsHandler *AnalyticsHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(LoggingMiddleware)

	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.CorsSettings.AllowedOrigins,
		AllowedMethods:   cfg.CorsSettings.AllowedMethods,
		AllowedHeaders:   cfg.CorsSettings.AllowedHeaders,
		AllowCredentials: true,
	})

	r.Use(c.Handler)

	r.Route("/api/v1", func(r chi.Router) {

		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projectHandler.GetAllProjects)
			r.Post("/", projectHandler.CreateProject)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", projectHandler.GetProjectByID)
				r.Put("/", projectHandler.UpdateProject)
				r.Delete("/", projectHandler.DeleteProject)

				r.Route("/analytics", func(r chi.Router) {
					r.Get("/", analyticsHandler.GetAnalytics)
					r.Post("/recalculate", analyticsHandler.Recalculate)
				})
			})
		})
	})

	return r
}
