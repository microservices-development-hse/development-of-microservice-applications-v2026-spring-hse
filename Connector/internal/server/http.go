package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/microservices-development-hse/connector/internal/handlers"
	jiraclient "github.com/microservices-development-hse/connector/internal/jira"
	"github.com/microservices-development-hse/connector/internal/logger"
)

type Server struct {
	httpServer *http.Server
}

func New(port int, client *jiraclient.Client, retryConfig jiraclient.RetryConfig, maxResults int, db *sql.DB, threadCount int) *Server {
	mux := http.NewServeMux()

	mux.Handle("/projects", handlers.NewProjectsHandler(client, retryConfig, maxResults, threadCount))
	mux.Handle("/updateProject", handlers.NewUpdateProjectHandler(client, retryConfig, maxResults, db, threadCount))

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      withMiddleware(mux),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 120 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	logger.Info("server: listening on %s", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server: listen: %w", err)
	}

	return nil
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("server: shutting down...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}

func withMiddleware(next http.Handler) http.Handler {
	return loggingMiddleware(recoveryMiddleware(corsMiddleware(next)))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logger.Info("server: --> %s %s", r.Method, r.URL.String())

		next.ServeHTTP(w, r)
		logger.Info("server: <-- %s %s (%s)", r.Method, r.URL.String(), time.Since(start))
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("server: panic recovered: %v", rec)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Для интеграционных тестов
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}
