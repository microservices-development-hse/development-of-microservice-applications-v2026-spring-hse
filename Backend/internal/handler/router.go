package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/microservices-development-hse/backend/internal/config"
	"github.com/rs/cors"
)

func NewRouter(cfg *config.Config, h *Handlers) *chi.Mux {
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
			r.Get("/", h.Project.GetAllProjects)
			r.Post("/", h.Project.CreateProject)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.Project.GetProjectByID)
				r.Put("/", h.Project.UpdateProject)
				r.Delete("/", h.Project.DeleteProject)

				r.Route("/analytics", func(r chi.Router) {
					r.Get("/", h.Analytics.GetAnalytics)
					r.Post("/recalculate", h.Analytics.Recalculate)
				})

				r.Get("/issues", h.Issue.GetProjectIssues)
			})
		})

		r.Route("/issues", func(r chi.Router) {
			r.Get("/{key}", h.Issue.GetIssueByKey)
		})

		r.Route("/sync", func(r chi.Router) {
			r.Post("/issue", h.Issue.SyncIssue)
		})
	})

	return r
}
