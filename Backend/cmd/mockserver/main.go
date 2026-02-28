package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/microservices-development-hse/backend/internal/handler"
	"github.com/microservices-development-hse/backend/internal/repository/mock"
	"github.com/microservices-development-hse/backend/internal/service"
)

func main() {
	repo := mock.NewAnalyticsMock()
	svc := service.NewAnalyticsService(repo)
	h := handler.NewHandler(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetProjects(w, r)
			return
		}
		if r.Method == http.MethodPost {
			h.AddProject(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			h.DeleteProject(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/stat") {
			h.ProjectStat(w, r)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/connector/updateProject", h.AddProject)
	mux.HandleFunc("/graph/make", h.MakeGraph)
	mux.HandleFunc("/graph/get", h.GetGraph)
	mux.HandleFunc("/graph/compare", h.CompareGraphs)
	mux.HandleFunc("/graph/delete", h.DeleteGraphs)

	mux.HandleFunc("/isAnalyzed", h.IsAnalyzed)
	mux.HandleFunc("/isEmpty", h.IsEmpty)

	handlerWithCORS := func(hh http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			hh.ServeHTTP(w, r)
		})
	}(mux)

	addr := ":8080"
	log.Printf("mock backend listening on %s\n", addr)
	if err := http.ListenAndServe(addr, handlerWithCORS); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
