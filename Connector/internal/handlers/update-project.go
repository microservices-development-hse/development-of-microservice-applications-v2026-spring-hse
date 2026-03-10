package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/microservices-development-hse/connector/internal/database"
	etlprocess "github.com/microservices-development-hse/connector/internal/etl-process"
	jiraclient "github.com/microservices-development-hse/connector/internal/jira"
	"github.com/microservices-development-hse/connector/internal/logger"
	dbmodels "github.com/microservices-development-hse/connector/internal/models/db"
)

type UpdateProjectHandler struct {
	extractor *etlprocess.Extractor
	loader    *etlprocess.Loader
}

func NewUpdateProjectHandler(client *jiraclient.Client, retryConfig jiraclient.RetryConfig, maxResults int, db *sql.DB) *UpdateProjectHandler {
	return &UpdateProjectHandler{
		extractor: etlprocess.NewExtractor(client, retryConfig, maxResults),
		loader: etlprocess.NewLoader(
			db,
			database.StmtUpsertProject,
			database.StmtUpsertIssue,
			database.StmtInsertUser,
		),
	}
}

func (h *UpdateProjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectKey := r.URL.Query().Get("project")
	if projectKey == "" {
		http.Error(w, "missing required parameter: project", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	logger.Info("updateProject: fetching issues for project %q", projectKey)

	jiraIssues, err := h.extractor.GetAllIssues(projectKey)
	if err != nil {
		logger.Error("updateProject: extract issues failed: %v", err)
		http.Error(w, "failed to fetch issues from Jira", http.StatusBadGateway)

		return
	}

	jiraProjects, err := h.extractor.GetProjects()
	if err != nil {
		logger.Error("updateProject: fetch projects failed: %v", err)
		http.Error(w, "failed to fetch project info from Jira", http.StatusBadGateway)

		return
	}

	var projectID int

	for _, jp := range jiraProjects {
		if jp.Key != projectKey {
			continue
		}

		dbProject, err := etlprocess.TransformProject(jp)
		if err != nil {
			logger.Error("updateProject: transform project: %v", err)
			http.Error(w, "failed to transform project", http.StatusInternalServerError)

			return
		}

		projectID, err = h.loader.LoadProject(ctx, dbProject)
		if err != nil {
			logger.Error("updateProject: load project: %v", err)
			http.Error(w, "failed to save project to database", http.StatusInternalServerError)

			return
		}

		break
	}

	if projectID == 0 {
		logger.Error("updateProject: project %q not found in Jira", projectKey)
		http.Error(w, "project not found in Jira", http.StatusNotFound)

		return
	}

	var allIssues []dbmodels.Issue

	var allUsers []dbmodels.User

	seenUsers := make(map[string]struct{})

	for _, ji := range jiraIssues {
		issue, users, err := etlprocess.TransformIssue(ji, projectID)
		if err != nil {
			logger.Error("updateProject: transform issue %q: %v", ji.Key, err)
			http.Error(w, "failed to transform issues", http.StatusInternalServerError)

			return
		}

		allIssues = append(allIssues, issue)

		for _, u := range users {
			if _, ok := seenUsers[u.Username]; ok {
				continue
			}

			seenUsers[u.Username] = struct{}{}
			allUsers = append(allUsers, u)
		}
	}

	if err := h.loader.LoadIssues(ctx, allIssues, allUsers); err != nil {
		logger.Error("updateProject: load issues: %v", err)
		http.Error(w, "failed to save issues to database", http.StatusInternalServerError)

		return
	}

	logger.Info("updateProject: project %q updated, %d issues loaded", projectKey, len(allIssues))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"project":     projectKey,
		"issuesCount": len(allIssues),
		"status":      "ok",
	})
}
