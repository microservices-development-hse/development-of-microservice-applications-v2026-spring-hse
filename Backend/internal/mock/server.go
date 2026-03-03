package mock

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"
	"strings"

	repoMock "github.com/microservices-development-hse/backend/internal/repository/mock"
)

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func NewMux() http.Handler {
	repo := repoMock.NewAnalyticsMock()
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("ok")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// GET /projects + POST /projects
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			q := r.URL.Query()

			page, _ := strconv.Atoi(q.Get("page"))
			if page <= 0 {
				page = 1
			}

			limit, _ := strconv.Atoi(q.Get("limit"))
			if limit <= 0 {
				limit = 20
			}

			search := q.Get("search")

			resp, err := repo.GetAllProjects(page, limit, search)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			writeJSON(w, http.StatusOK, resp)

			return
		}

		if r.Method == http.MethodPost {
			var payload struct {
				Key string `json:"key"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}

			res, err := repo.AddProjectFromJira(payload.Key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			writeJSON(w, http.StatusOK, res)

			return
		}

		http.NotFound(w, r)
	})

	// DELETE /projects/{id}  +  GET /projects/{id}/stat
	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			idStr := path.Base(r.URL.Path)

			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}

			res, err := repo.DeleteProjectByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			writeJSON(w, http.StatusOK, res)

			return
		}

		if strings.HasSuffix(r.URL.Path, "/stat") {
			parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(parts) < 2 {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}

			id := parts[1]

			stat, err := repo.GetProjectStatByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			writeJSON(w, http.StatusOK, stat)

			return
		}

		http.NotFound(w, r)
	})

	// POST /connector/updateProject
	mux.HandleFunc("/connector/updateProject", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Key string `json:"key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		res, err := repo.AddProjectFromJira(payload.Key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, res)
	})

	// Graph endpoints
	mux.HandleFunc("/graph/make", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Task    string `json:"task"`
			Project string `json:"project"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		job, err := repo.MakeGraph(payload.Task, payload.Project)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, job)
	})

	mux.HandleFunc("/graph/get", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		task := q.Get("task")
		project := q.Get("project")

		res, err := repo.GetGraph(task, project)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, res)
	})

	mux.HandleFunc("/graph/compare", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Task     string   `json:"task"`
			Projects []string `json:"projects"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		res, err := repo.CompareGraphs(payload.Task, payload.Projects)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, res)
	})

	mux.HandleFunc("/graph/delete", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Project string `json:"project"`
		}

		_ = json.NewDecoder(r.Body).Decode(&payload)

		res, err := repo.DeleteGraphs(payload.Project)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, res)
	})

	// Checks
	mux.HandleFunc("/isAnalyzed", func(w http.ResponseWriter, r *http.Request) {
		project := r.URL.Query().Get("project")

		ok, err := repo.IsAnalyzed(project)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]bool{"analyzed": ok})
	})

	mux.HandleFunc("/isEmpty", func(w http.ResponseWriter, r *http.Request) {
		project := r.URL.Query().Get("project")

		ok, err := repo.IsEmpty(project)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]bool{"empty": ok})
	})

	return enableCORS(mux)
}
