package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	dbmodels "github.com/microservices-development-hse/connector/internal/models/db"
	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
)

//
// ===================== MOCKS =====================
//

type mockExtractor struct {
	projects []jiramodels.ProjectResponse
	issues   []jiramodels.Issue
	err      error
}

func (m *mockExtractor) GetProjects() ([]jiramodels.ProjectResponse, error) {
	return m.projects, m.err
}

func (m *mockExtractor) GetAllIssues(ctx context.Context, projectKey string) ([]jiramodels.Issue, error) {
	return m.issues, m.err
}

type mockLoader struct {
	projectID int
	authorIDs map[string]int
	issueIDs  map[string]int
	err       error
}

func (m *mockLoader) LoadProject(ctx context.Context, p dbmodels.Project) (int, error) {
	return m.projectID, m.err
}

func (m *mockLoader) UpsertAuthors(ctx context.Context, a map[string]dbmodels.Author) (map[string]int, error) {
	return m.authorIDs, m.err
}

func (m *mockLoader) LoadIssues(ctx context.Context, issues []dbmodels.Issue) (map[string]int, error) {
	return m.issueIDs, m.err
}

func (m *mockLoader) LoadStatusChanges(ctx context.Context, changes []dbmodels.StatusChange) error {
	return m.err
}

//
// ===================== PROJECTS HANDLER =====================
//

func TestProjectsHandler_MethodNotAllowed(t *testing.T) {
	h := NewProjectsHandlerWithExtractor(&mockExtractor{})

	req := httptest.NewRequest(http.MethodPost, "/projects", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestProjectsHandler_Success(t *testing.T) {
	extractor := &mockExtractor{
		projects: []jiramodels.ProjectResponse{
			{Key: "A", Name: "Alpha", Self: "url1"},
			{Key: "B", Name: "Beta", Self: "url2"},
		},
	}

	h := NewProjectsHandlerWithExtractor(extractor)

	req := httptest.NewRequest(http.MethodGet, "/projects?limit=1&page=1", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ProjectsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if len(resp.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(resp.Projects))
	}
}

func TestProjectsHandler_SearchFilter(t *testing.T) {
	extractor := &mockExtractor{
		projects: []jiramodels.ProjectResponse{
			{Key: "TEST", Name: "Alpha", Self: "url"},
			{Key: "XYZ", Name: "Beta", Self: "url"},
		},
	}

	h := NewProjectsHandlerWithExtractor(extractor)

	req := httptest.NewRequest(http.MethodGet, "/projects?search=test", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	var resp ProjectsResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if len(resp.Projects) != 1 {
		t.Fatalf("expected filtered result")
	}
}

func TestProjectsHandler_ExtractorError(t *testing.T) {
	extractor := &mockExtractor{err: errors.New("fail")}
	h := NewProjectsHandlerWithExtractor(extractor)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", w.Code)
	}
}

//
// ===================== UPDATE PROJECT HANDLER =====================
//

func TestUpdateProject_MethodNotAllowed(t *testing.T) {
	h := NewUpdateProjectHandlerWithDeps(&mockExtractor{}, &mockLoader{})

	req := httptest.NewRequest(http.MethodGet, "/updateProject", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405")
	}
}

func TestUpdateProject_MissingParam(t *testing.T) {
	h := NewUpdateProjectHandlerWithDeps(&mockExtractor{}, &mockLoader{})

	req := httptest.NewRequest(http.MethodPost, "/updateProject", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}
}

func TestUpdateProject_ExtractorError(t *testing.T) {
	h := NewUpdateProjectHandlerWithDeps(
		&mockExtractor{err: errors.New("fail")},
		&mockLoader{},
	)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502")
	}
}

func TestUpdateProject_ProjectNotFound(t *testing.T) {
	extractor := &mockExtractor{
		projects: []jiramodels.ProjectResponse{},
		issues:   []jiramodels.Issue{},
	}

	h := NewUpdateProjectHandlerWithDeps(extractor, &mockLoader{})

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404")
	}
}

func TestUpdateProject_NoIssues(t *testing.T) {
	extractor := &mockExtractor{
		projects: []jiramodels.ProjectResponse{
			{ID: "1", Key: "X", Name: "Test"},
		},
		issues: []jiramodels.Issue{},
	}

	loader := &mockLoader{projectID: 1}

	h := NewUpdateProjectHandlerWithDeps(extractor, loader)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}

func TestUpdateProject_FullSuccess(t *testing.T) {
	extractor := &mockExtractor{
		projects: []jiramodels.ProjectResponse{
			{ID: "1", Key: "X", Name: "Test"},
		},
		issues: []jiramodels.Issue{
			{
				ID:  "1",
				Key: "X-1",
				Fields: jiramodels.Fields{
					Summary: "test",
					Status:  jiramodels.Status{Name: "Open"},
					Priority: jiramodels.Priority{
						Name: "High",
					},
				},
			},
		},
	}

	loader := &mockLoader{
		projectID: 1,
		authorIDs: map[string]int{},
		issueIDs:  map[string]int{"X-1": 1},
	}

	h := NewUpdateProjectHandlerWithDeps(extractor, loader)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
