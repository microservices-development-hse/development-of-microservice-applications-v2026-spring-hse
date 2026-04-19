package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
)

func TestUpdateProject_MethodNotAllowed(t *testing.T) {
	env := SetupTestEnv(t, mockJiraHandler(nil, nil))

	resp := env.GET(t, "/updateProject?project=AAR")

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestUpdateProject_MissingProjectParam(t *testing.T) {
	env := SetupTestEnv(t, mockJiraHandler(nil, nil))

	resp := env.POST(t, "/updateProject")

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateProject_ProjectNotFoundInJira(t *testing.T) {
	env := SetupTestEnv(t, mockJiraHandler(
		[]jiramodels.ProjectResponse{},
		[]jiramodels.Issue{makeJiraIssue("1", "TEST-1", false)},
	))

	resp := env.POST(t, "/updateProject?project=NOTEXIST")

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdateProject_FullCycle(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "TEST", "Test Project"),
	}

	issues := []jiramodels.Issue{
		makeJiraIssue("10001", "TEST-1", true),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, issues))

	resp := env.POST(t, "/updateProject?project=TEST")

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", body["status"])
	}

	if int(body["issuesCount"].(float64)) != 1 {
		t.Errorf("expected issuesCount=1, got %v", body["issuesCount"])
	}

	if count := env.CountRows(t, "projects"); count != 1 {
		t.Errorf("expected 1 project in DB, got %d", count)
	}

	if count := env.CountRows(t, "issues"); count != 1 {
		t.Errorf("expected 1 issue in DB, got %d", count)
	}

	if count := env.CountRows(t, "authors"); count != 2 {
		t.Errorf("expected 2 authors (creator + assignee), got %d", count)
	}

	if count := env.CountRows(t, "status_changes"); count < 1 {
		t.Errorf("expected at least 1 status_change in DB, got %d", count)
	}
}

func TestUpdateProject_EmptyProject(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "EMPTY", "Empty Project"),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, []jiramodels.Issue{}))

	resp := env.POST(t, "/updateProject?project=EMPTY")

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if int(body["issuesCount"].(float64)) != 0 {
		t.Errorf("expected issuesCount=0, got %v", body["issuesCount"])
	}

	if count := env.CountRows(t, "projects"); count != 1 {
		t.Errorf("expected 1 project in DB even with 0 issues, got %d", count)
	}
}

func TestUpdateProject_MultipleIssues(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "MULTI", "Multi Project"),
	}

	issues := []jiramodels.Issue{
		makeJiraIssue("1", "MULTI-1", false),
		makeJiraIssue("2", "MULTI-2", false),
		makeJiraIssue("3", "MULTI-3", true),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, issues))

	resp := env.POST(t, "/updateProject?project=MULTI")

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if count := env.CountRows(t, "issues"); count != 3 {
		t.Errorf("expected 3 issues in DB, got %d", count)
	}
}

func TestUpdateProject_Idempotent(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "TEST", "Test Project"),
	}

	issues := []jiramodels.Issue{
		makeJiraIssue("10001", "TEST-1", true),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, issues))

	resp1 := env.POST(t, "/updateProject?project=TEST")

	defer func() {
		_ = resp1.Body.Close()
	}()

	resp2 := env.POST(t, "/updateProject?project=TEST")

	defer func() {
		_ = resp2.Body.Close()
	}()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on second call, got %d", resp2.StatusCode)
	}

	if count := env.CountRows(t, "projects"); count != 1 {
		t.Errorf("expected 1 project after 2 calls, got %d", count)
	}

	if count := env.CountRows(t, "issues"); count != 1 {
		t.Errorf("expected 1 issue after 2 calls (upsert), got %d", count)
	}
}

func TestUpdateProject_IssueDataCorrect(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "CHECK", "Check Project"),
	}

	issues := []jiramodels.Issue{
		makeJiraIssue("42", "CHECK-1", false),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, issues))

	resp := env.POST(t, "/updateProject?project=CHECK")

	defer func() {
		_ = resp.Body.Close()
	}()

	var key, summary, status, priority, externalID string

	err := env.DB.QueryRow(`
		SELECT key, summary, status, priority, external_id
		FROM issues LIMIT 1
	`).Scan(&key, &summary, &status, &priority, &externalID)
	if err != nil {
		t.Fatalf("query issue failed: %v", err)
	}

	if key != "CHECK-1" {
		t.Errorf("expected key=CHECK-1, got %s", key)
	}

	if summary != "Summary for CHECK-1" {
		t.Errorf("expected summary, got %s", summary)
	}

	if status != "Open" {
		t.Errorf("expected status=Open, got %s", status)
	}

	if priority != "High" {
		t.Errorf("expected priority=High, got %s", priority)
	}

	if externalID != "42" {
		t.Errorf("expected external_id=42, got %s", externalID)
	}
}

func TestUpdateProject_StatusChangesCorrect(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "SC", "Status Change Project"),
	}

	issues := []jiramodels.Issue{
		makeJiraIssue("1", "SC-1", true),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, issues))

	resp := env.POST(t, "/updateProject?project=SC")

	defer func() {
		_ = resp.Body.Close()
	}()

	var fromStatus, toStatus string

	err := env.DB.QueryRow(`
		SELECT from_status, to_status FROM status_changes LIMIT 1
	`).Scan(&fromStatus, &toStatus)
	if err != nil {
		t.Fatalf("query status_change failed: %v", err)
	}

	if fromStatus != "Open" {
		t.Errorf("expected from_status=Open, got %s", fromStatus)
	}

	if toStatus != "In Progress" {
		t.Errorf("expected to_status=In Progress, got %s", toStatus)
	}
}

func TestUpdateProject_JiraError(t *testing.T) {
	env := SetupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))

	resp := env.POST(t, "/updateProject?project=TEST")

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", resp.StatusCode)
	}
}

func TestUpdateProject_TransactionAtomicity_UpsertScenario(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "ATOM", "Atomic Project"),
	}

	issues := []jiramodels.Issue{
		makeJiraIssue("ID-1", "ATOM-1", false),
		makeJiraIssue("ID-1", "ATOM-1", false),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, issues))

	resp := env.POST(t, "/updateProject?project=ATOM")

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if count := env.CountRows(t, "issues"); count != 1 {
		t.Errorf("expected 1 issue after upsert of duplicate, got %d", count)
	}
}

func TestUpdateProject_TransactionAtomicity_RollbackScenario(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "ROLL", "Rollback Project"),
	}

	issues := []jiramodels.Issue{
		makeJiraIssue("1", "ROLL-1", false),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, issues))

	resp := env.POST(t, "/updateProject?project=ROLL")

	defer func() {
		_ = resp.Body.Close()
	}()

	countBefore := env.CountRows(t, "issues")

	resp2 := env.POST(t, "/updateProject?project=ROLL")

	defer func() {
		_ = resp2.Body.Close()
	}()

	countAfter := env.CountRows(t, "issues")

	if countBefore != countAfter {
		t.Errorf("repeated load changed issue count: before=%d, after=%d", countBefore, countAfter)
	}
}

func TestUpdateProject_MultipleProjects(t *testing.T) {
	allProjects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "PROJ1", "Project One"),
		makeJiraProject("2", "PROJ2", "Project Two"),
	}

	issues1 := []jiramodels.Issue{makeJiraIssue("1", "PROJ1-1", false)}
	issues2 := []jiramodels.Issue{makeJiraIssue("2", "PROJ2-1", false)}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/rest/api/2/project" {
			_ = json.NewEncoder(w).Encode(allProjects)

			return
		}

		jql := r.URL.Query().Get("jql")
		maxResults := r.URL.Query().Get("maxResults")

		var issuesForProject []jiramodels.Issue

		if strings.Contains(jql, "PROJ1") {
			issuesForProject = issues1
		} else {
			issuesForProject = issues2
		}

		if maxResults == "1" {
			total := len(issuesForProject)

			var first []jiramodels.Issue

			if total > 0 {
				first = issuesForProject[:1]
			}

			_ = json.NewEncoder(w).Encode(jiramodels.IssueSearchResponse{
				Total: total, Issues: first,
			})

			return
		}

		_ = json.NewEncoder(w).Encode(jiramodels.IssueSearchResponse{
			Total: len(issuesForProject), Issues: issuesForProject,
		})
	})

	env := SetupTestEnv(t, handler)

	resp1 := env.POST(t, "/updateProject?project=PROJ1")

	defer func() {
		_ = resp1.Body.Close()
	}()

	if resp1.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for PROJ1, got %d", resp1.StatusCode)
	}

	resp2 := env.POST(t, "/updateProject?project=PROJ2")

	defer func() {
		_ = resp2.Body.Close()
	}()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for PROJ2, got %d", resp2.StatusCode)
	}

	if count := env.CountRows(t, "projects"); count != 2 {
		t.Errorf("expected 2 projects in DB, got %d", count)
	}

	if count := env.CountRows(t, "issues"); count != 2 {
		t.Errorf("expected 2 issues in DB, got %d", count)
	}
}
