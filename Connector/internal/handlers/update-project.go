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

func NewUpdateProjectHandler(client *jiraclient.Client, retryConfig jiraclient.RetryConfig, maxResults int, db *sql.DB, threadCount int) *UpdateProjectHandler {
	return &UpdateProjectHandler{
		extractor: etlprocess.NewExtractor(
			client,
			retryConfig,
			maxResults,
			threadCount,
		),
		loader: etlprocess.NewLoader(
			db,
			database.StmtUpsertProject,
			database.StmtUpsertIssue,
			database.StmtInsertAuthor,
			database.StmtInsertStatusChange,
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

	jiraIssues, err := h.extractor.GetAllIssues(ctx, projectKey)
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

	seenAuthors := make(map[string]dbmodels.Author)

	for _, ji := range jiraIssues {
		if ji.Fields.Creator.Name != "" {
			seenAuthors[ji.Fields.Creator.Name] = dbmodels.Author{
				ExternalID: ji.Fields.Creator.Name,
				Username:   ji.Fields.Creator.DisplayName,
			}
		}

		if ji.Fields.Assignee != nil && ji.Fields.Assignee.Name != "" {
			seenAuthors[ji.Fields.Assignee.Name] = dbmodels.Author{
				ExternalID: ji.Fields.Assignee.Name,
				Username:   ji.Fields.Assignee.DisplayName,
			}
		}

		if ji.Changelog != nil {
			for _, h := range ji.Changelog.Histories {
				if h.Author.Name != "" {
					seenAuthors[h.Author.Name] = dbmodels.Author{
						ExternalID: h.Author.Name,
						Username:   h.Author.DisplayName,
					}
				}
			}
		}
	}

	authorIDs, err := h.loader.UpsertAuthors(ctx, seenAuthors)
	if err != nil {
		logger.Error("updateProject: upsert authors: %v", err)
		http.Error(w, "failed to save authors to database", http.StatusInternalServerError)

		return
	}

	var allIssues []dbmodels.Issue

	for _, ji := range jiraIssues {
		var authorID *int
		if id := authorIDs[ji.Fields.Creator.Name]; id != 0 {
			authorID = &id
		}

		var assigneeID *int

		if ji.Fields.Assignee != nil && ji.Fields.Assignee.Name != "" {
			if id := authorIDs[ji.Fields.Assignee.Name]; id != 0 {
				assigneeID = &id
			}
		}

		issue, err := etlprocess.TransformIssue(ji, projectID, authorID, assigneeID)
		if err != nil {
			logger.Error("updateProject: transform issue %q: %v", ji.Key, err)
			http.Error(w, "failed to transform issues", http.StatusInternalServerError)

			return
		}

		allIssues = append(allIssues, issue)
	}

	if len(allIssues) == 0 {
		logger.Info("updateProject: no issues found for project %q", projectKey)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"project":     projectKey,
			"issuesCount": 0,
			"status":      "ok",
		})

		return
	}

	issueIDs, err := h.loader.LoadIssues(ctx, allIssues)
	if err != nil {
		logger.Error("updateProject: load issues: %v", err)
		http.Error(w, "failed to save issues to database", http.StatusInternalServerError)

		return
	}

	var allStatusChanges []dbmodels.StatusChange

	for _, ji := range jiraIssues {
		if ji.Changelog == nil {
			continue
		}

		issueID, ok := issueIDs[ji.Key]
		if !ok {
			continue
		}

		changes := etlprocess.TransformStatusChanges(ji.Changelog, issueID, authorIDs)
		allStatusChanges = append(allStatusChanges, changes...)
	}

	if err := h.loader.LoadStatusChanges(ctx, allStatusChanges); err != nil {
		logger.Error("updateProject: load status changes: %v", err)
		http.Error(w, "failed to save status changes to database", http.StatusInternalServerError)

		return
	}

	logger.Info("updateProject: project %q updated, %d issues, %d status changes loaded",
		projectKey, len(allIssues), len(allStatusChanges))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"project":      projectKey,
		"issuesCount":  len(allIssues),
		"changesCount": len(allStatusChanges),
		"status":       "ok",
	})
}
