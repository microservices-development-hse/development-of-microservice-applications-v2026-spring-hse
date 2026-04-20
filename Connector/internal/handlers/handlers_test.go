package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// ===================== ДОПОЛНИТЕЛЬНЫЕ МОКИ =====================

type mockExtractorWithProjects struct {
	projects    []jiramodels.ProjectResponse
	issues      []jiramodels.Issue
	err         error
	projectsErr error
}

func (m *mockExtractorWithProjects) GetProjects() ([]jiramodels.ProjectResponse, error) {
	if m.projectsErr != nil {
		return nil, m.projectsErr
	}

	return m.projects, nil
}

func (m *mockExtractorWithProjects) GetAllIssues(ctx context.Context, projectKey string) ([]jiramodels.Issue, error) {
	return m.issues, m.err
}

type mockLoaderWithError struct {
	projectID            int
	authorIDs            map[string]int
	issueIDs             map[string]int
	err                  error
	loadProjectErr       error
	upsertAuthorsErr     error
	loadIssuesErr        error
	loadStatusChangesErr error
}

func (m *mockLoaderWithError) LoadProject(ctx context.Context, p dbmodels.Project) (int, error) {
	if m.loadProjectErr != nil {
		return 0, m.loadProjectErr
	}

	return m.projectID, m.err
}

func (m *mockLoaderWithError) UpsertAuthors(ctx context.Context, a map[string]dbmodels.Author) (map[string]int, error) {
	if m.upsertAuthorsErr != nil {
		return nil, m.upsertAuthorsErr
	}

	return m.authorIDs, m.err
}

func (m *mockLoaderWithError) LoadIssues(ctx context.Context, issues []dbmodels.Issue) (map[string]int, error) {
	if m.loadIssuesErr != nil {
		return nil, m.loadIssuesErr
	}

	return m.issueIDs, m.err
}

func (m *mockLoaderWithError) LoadStatusChanges(ctx context.Context, changes []dbmodels.StatusChange) error {
	if m.loadStatusChangesErr != nil {
		return m.loadStatusChangesErr
	}

	return m.err
}

// ===================== ТЕСТЫ ДЛЯ PROJECTS HANDLER =====================

func TestProjectsHandler_Pagination_StartGreaterThanTotal(t *testing.T) {
	extractor := &mockExtractor{
		projects: []jiramodels.ProjectResponse{
			{Key: "A", Name: "Alpha", Self: "url1"},
			{Key: "B", Name: "Beta", Self: "url2"},
		},
	}

	h := NewProjectsHandlerWithExtractor(extractor)

	req := httptest.NewRequest(http.MethodGet, "/projects?limit=1&page=100", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ProjectsResponse

	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if len(resp.Projects) != 0 {
		t.Errorf("expected 0 projects when start > total, got %d", len(resp.Projects))
	}
}

func TestProjectsHandler_Pagination_ZeroTotal(t *testing.T) {
	extractor := &mockExtractor{
		projects: []jiramodels.ProjectResponse{},
	}

	h := NewProjectsHandlerWithExtractor(extractor)

	req := httptest.NewRequest(http.MethodGet, "/projects?limit=10&page=1", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ProjectsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp.PageInfo.PageCount != 1 {
		t.Errorf("expected pageCount=1 when total=0, got %d", resp.PageInfo.PageCount)
	}
}

func TestProjectsHandler_EncodeError(t *testing.T) {
	extractor := &mockExtractor{
		projects: []jiramodels.ProjectResponse{
			{Key: "A", Name: "Alpha", Self: "url1"},
		},
	}

	h := NewProjectsHandlerWithExtractor(extractor)

	req := httptest.NewRequest(http.MethodGet, "/projects?limit=1&page=1", nil)
	w := &errorResponseWriter{httptest.NewRecorder(), true}

	h.ServeHTTP(w, req)
}

type errorResponseWriter struct {
	*httptest.ResponseRecorder
	forceError bool
}

func (e *errorResponseWriter) Write(b []byte) (int, error) {
	if e.forceError {
		return 0, fmt.Errorf("write error: %w", errors.New("forced error"))
	}

	n, err := e.ResponseRecorder.Write(b)
	if err != nil {
		return n, fmt.Errorf("response write failed: %w", err)
	}

	return n, nil
}

// ===================== ТЕСТЫ ДЛЯ UPDATE PROJECT HANDLER =====================

func TestUpdateProject_GetProjectsError(t *testing.T) {
	extractor := &mockExtractorWithProjects{
		issues: []jiramodels.Issue{
			{ID: "1", Key: "X-1", Fields: jiramodels.Fields{Creator: jiramodels.Author{Name: "user1"}}},
		},
		projectsErr: errors.New("failed to get projects"),
	}
	loader := &mockLoader{}

	h := NewUpdateProjectHandlerWithDeps(extractor, loader)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", w.Code)
	}
}

func TestUpdateProject_TransformProjectError(t *testing.T) {
	extractor := &mockExtractorWithProjects{
		projects: []jiramodels.ProjectResponse{
			{ID: "invalid", Key: "X", Name: "Test"},
		},
		issues: []jiramodels.Issue{},
	}
	loader := &mockLoader{}

	h := NewUpdateProjectHandlerWithDeps(extractor, loader)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUpdateProject_LoadProjectError(t *testing.T) {
	extractor := &mockExtractorWithProjects{
		projects: []jiramodels.ProjectResponse{
			{ID: "1", Key: "X", Name: "Test"},
		},
		issues: []jiramodels.Issue{},
	}
	loader := &mockLoaderWithError{
		loadProjectErr: errors.New("db error"),
	}

	h := NewUpdateProjectHandlerWithDeps(extractor, loader)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUpdateProject_UpsertAuthorsError(t *testing.T) {
	extractor := &mockExtractorWithProjects{
		projects: []jiramodels.ProjectResponse{
			{ID: "1", Key: "X", Name: "Test"},
		},
		issues: []jiramodels.Issue{
			{
				ID:  "1",
				Key: "X-1",
				Fields: jiramodels.Fields{
					Creator: jiramodels.Author{Name: "user1", DisplayName: "User One"},
				},
			},
		},
	}
	loader := &mockLoaderWithError{
		projectID:        1,
		upsertAuthorsErr: errors.New("upsert authors failed"),
	}

	h := NewUpdateProjectHandlerWithDeps(extractor, loader)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUpdateProject_TransformIssueError(t *testing.T) {
	extractor := &mockExtractorWithProjects{
		projects: []jiramodels.ProjectResponse{
			{ID: "invalid", Key: "X", Name: "Test"},
		},
		issues: []jiramodels.Issue{
			{
				ID:  "invalid",
				Key: "X-1",
				Fields: jiramodels.Fields{
					Creator: jiramodels.Author{Name: "user1"},
				},
			},
		},
	}
	loader := &mockLoaderWithError{
		projectID: 1,
		authorIDs: map[string]int{"user1": 1},
	}

	h := NewUpdateProjectHandlerWithDeps(extractor, loader)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUpdateProject_LoadIssuesError(t *testing.T) {
	extractor := &mockExtractorWithProjects{
		projects: []jiramodels.ProjectResponse{
			{ID: "1", Key: "X", Name: "Test"},
		},
		issues: []jiramodels.Issue{
			{
				ID:  "1",
				Key: "X-1",
				Fields: jiramodels.Fields{
					Creator:  jiramodels.Author{Name: "user1", DisplayName: "User One"},
					Status:   jiramodels.Status{Name: "Open"},
					Priority: jiramodels.Priority{Name: "High"},
				},
			},
		},
	}
	loader := &mockLoaderWithError{
		projectID:     1,
		authorIDs:     map[string]int{"user1": 1},
		loadIssuesErr: errors.New("load issues failed"),
	}

	h := NewUpdateProjectHandlerWithDeps(extractor, loader)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUpdateProject_LoadStatusChangesError(t *testing.T) {
	extractor := &mockExtractorWithProjects{
		projects: []jiramodels.ProjectResponse{
			{ID: "1", Key: "X", Name: "Test"},
		},
		issues: []jiramodels.Issue{
			{
				ID:  "1",
				Key: "X-1",
				Fields: jiramodels.Fields{
					Creator:  jiramodels.Author{Name: "user1", DisplayName: "User One"},
					Status:   jiramodels.Status{Name: "Open"},
					Priority: jiramodels.Priority{Name: "High"},
				},
				Changelog: &jiramodels.Changelog{
					Histories: []jiramodels.History{
						{
							Author:  jiramodels.Author{Name: "user1"},
							Created: jiramodels.JTime{},
							Items:   []jiramodels.Item{{Field: "status", From: "Open", To: "Closed"}},
						},
					},
				},
			},
		},
	}
	loader := &mockLoaderWithError{
		projectID:            1,
		authorIDs:            map[string]int{"user1": 1},
		issueIDs:             map[string]int{"X-1": 1},
		loadStatusChangesErr: errors.New("load status changes failed"),
	}

	h := NewUpdateProjectHandlerWithDeps(extractor, loader)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUpdateProject_WithAssigneeAndChangelog(t *testing.T) {
	extractor := &mockExtractorWithProjects{
		projects: []jiramodels.ProjectResponse{
			{ID: "1", Key: "X", Name: "Test"},
		},
		issues: []jiramodels.Issue{
			{
				ID:  "1",
				Key: "X-1",
				Fields: jiramodels.Fields{
					Creator:  jiramodels.Author{Name: "user1", DisplayName: "User One"},
					Assignee: &jiramodels.Author{Name: "user2", DisplayName: "User Two"},
					Status:   jiramodels.Status{Name: "Open"},
					Priority: jiramodels.Priority{Name: "High"},
				},
				Changelog: &jiramodels.Changelog{
					Histories: []jiramodels.History{
						{
							Author:  jiramodels.Author{Name: "user2"},
							Created: jiramodels.JTime{},
							Items:   []jiramodels.Item{{Field: "status", From: "Open", To: "In Progress"}},
						},
					},
				},
			},
		},
	}
	loader := &mockLoaderWithError{
		projectID: 1,
		authorIDs: map[string]int{"user1": 1, "user2": 2},
		issueIDs:  map[string]int{"X-1": 1},
	}

	h := NewUpdateProjectHandlerWithDeps(extractor, loader)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp["status"])
	}
}

func TestUpdateProject_ChangelogWithoutIssueID(t *testing.T) {
	extractor := &mockExtractorWithProjects{
		projects: []jiramodels.ProjectResponse{
			{ID: "1", Key: "X", Name: "Test"},
		},
		issues: []jiramodels.Issue{
			{
				ID:  "1",
				Key: "X-1",
				Fields: jiramodels.Fields{
					Creator:  jiramodels.Author{Name: "user1", DisplayName: "User One"},
					Status:   jiramodels.Status{Name: "Open"},
					Priority: jiramodels.Priority{Name: "High"},
				},
				Changelog: &jiramodels.Changelog{
					Histories: []jiramodels.History{
						{
							Author:  jiramodels.Author{Name: "user1"},
							Created: jiramodels.JTime{},
							Items:   []jiramodels.Item{{Field: "status", From: "Open", To: "Closed"}},
						},
					},
				},
			},
		},
	}
	loader := &mockLoaderWithError{
		projectID: 1,
		authorIDs: map[string]int{"user1": 1},
		issueIDs:  map[string]int{},
	}

	h := NewUpdateProjectHandlerWithDeps(extractor, loader)

	req := httptest.NewRequest(http.MethodPost, "/updateProject?project=X", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
