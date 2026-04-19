package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
)

func TestProjects_MethodNotAllowed(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "AAR", "aardvark"),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, nil))

	resp, err := http.Post(env.HTTPServer.URL+"/projects", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestProjects_ReturnsAllProjects(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "AAR", "aardvark"),
		makeJiraProject("2", "AVRO", "Avro"),
		makeJiraProject("3", "ACE", "ACE"),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, nil))

	resp := env.GET(t, "/projects")

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Projects []ProjectItem `json:"projects"`
		PageInfo PageInfo      `json:"pageInfo"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if body.PageInfo.ProjectsCount != 3 {
		t.Errorf("expected 3 projects, got %d", body.PageInfo.ProjectsCount)
	}

	if len(body.Projects) != 3 {
		t.Errorf("expected 3 projects in response, got %d", len(body.Projects))
	}
}

func TestProjects_Pagination(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "AAR", "aardvark"),
		makeJiraProject("2", "AVRO", "Avro"),
		makeJiraProject("3", "ACE", "ACE"),
		makeJiraProject("4", "AMQ", "ActiveMQ"),
		makeJiraProject("5", "ABDERA", "Abdera"),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, nil))

	resp := env.GET(t, "/projects?limit=2&page=1")

	defer func() {
		_ = resp.Body.Close()
	}()

	var body struct {
		Projects []ProjectItem `json:"projects"`
		PageInfo PageInfo      `json:"pageInfo"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(body.Projects) != 2 {
		t.Errorf("expected 2 projects on page 1, got %d", len(body.Projects))
	}

	if body.PageInfo.PageCount != 3 {
		t.Errorf("expected pageCount=3, got %d", body.PageInfo.PageCount)
	}

	if body.PageInfo.CurrentPage != 1 {
		t.Errorf("expected currentPage=1, got %d", body.PageInfo.CurrentPage)
	}

	if body.PageInfo.ProjectsCount != 5 {
		t.Errorf("expected projectsCount=5, got %d", body.PageInfo.ProjectsCount)
	}

	resp2 := env.GET(t, "/projects?limit=2&page=2")

	defer func() {
		_ = resp2.Body.Close()
	}()

	var body2 struct {
		Projects []ProjectItem `json:"projects"`
		PageInfo PageInfo      `json:"pageInfo"`
	}

	if err := json.NewDecoder(resp2.Body).Decode(&body2); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(body2.Projects) != 2 {
		t.Errorf("expected 2 projects on page 2, got %d", len(body2.Projects))
	}

	if body2.PageInfo.CurrentPage != 2 {
		t.Errorf("expected currentPage=2, got %d", body2.PageInfo.CurrentPage)
	}

	resp3 := env.GET(t, "/projects?limit=2&page=3")

	defer func() {
		_ = resp3.Body.Close()
	}()

	var body3 struct {
		Projects []ProjectItem `json:"projects"`
		PageInfo PageInfo      `json:"pageInfo"`
	}
	if err := json.NewDecoder(resp3.Body).Decode(&body3); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(body3.Projects) != 1 {
		t.Errorf("expected 1 project on last page, got %d", len(body3.Projects))
	}
}

func TestProjects_SearchByKey(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "AAR", "aardvark"),
		makeJiraProject("2", "AVRO", "Avro"),
		makeJiraProject("3", "AMQ", "ActiveMQ"),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, nil))

	resp := env.GET(t, "/projects?search=avro")

	defer func() {
		_ = resp.Body.Close()
	}()

	var body struct {
		Projects []ProjectItem `json:"projects"`
		PageInfo PageInfo      `json:"pageInfo"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if body.PageInfo.ProjectsCount != 1 {
		t.Errorf("expected 1 project matching 'avro', got %d", body.PageInfo.ProjectsCount)
	}

	if body.Projects[0].Key != "AVRO" {
		t.Errorf("expected AVRO, got %s", body.Projects[0].Key)
	}
}

func TestProjects_SearchByName(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "AAR", "aardvark"),
		makeJiraProject("2", "AVRO", "Avro"),
		makeJiraProject("3", "AMQ", "ActiveMQ"),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, nil))

	resp := env.GET(t, "/projects?search=active")

	defer func() {
		_ = resp.Body.Close()
	}()

	var body struct {
		Projects []ProjectItem `json:"projects"`
		PageInfo PageInfo      `json:"pageInfo"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if body.PageInfo.ProjectsCount != 1 {
		t.Errorf("expected 1 project matching 'active', got %d", body.PageInfo.ProjectsCount)
	}

	if body.Projects[0].Key != "AMQ" {
		t.Errorf("expected AMQ, got %s", body.Projects[0].Key)
	}
}

func TestProjects_SearchCaseInsensitive(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "AAR", "Aardvark"),
		makeJiraProject("2", "AVRO", "Avro"),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, nil))

	resp := env.GET(t, "/projects?search=AARDVARK")

	defer func() {
		_ = resp.Body.Close()
	}()

	var body struct {
		Projects []ProjectItem `json:"projects"`
		PageInfo PageInfo      `json:"pageInfo"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if body.PageInfo.ProjectsCount != 1 {
		t.Errorf("expected 1 project for case-insensitive search, got %d", body.PageInfo.ProjectsCount)
	}
}

func TestProjects_SearchNoResults(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "AAR", "aardvark"),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, nil))

	resp := env.GET(t, "/projects?search=nonexistent")

	defer func() {
		_ = resp.Body.Close()
	}()

	var body struct {
		Projects []ProjectItem `json:"projects"`
		PageInfo PageInfo      `json:"pageInfo"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if body.PageInfo.ProjectsCount != 0 {
		t.Errorf("expected 0 projects, got %d", body.PageInfo.ProjectsCount)
	}

	if body.PageInfo.PageCount != 1 {
		t.Errorf("expected pageCount=1 for empty result, got %d", body.PageInfo.PageCount)
	}
}

func TestProjects_PageBeyondTotal(t *testing.T) {
	projects := []jiramodels.ProjectResponse{
		makeJiraProject("1", "AAR", "aardvark"),
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, nil))

	resp := env.GET(t, "/projects?limit=10&page=99")

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Projects []ProjectItem `json:"projects"`
		PageInfo PageInfo      `json:"pageInfo"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(body.Projects) != 0 {
		t.Errorf("expected 0 projects for page beyond total, got %d", len(body.Projects))
	}
}

func TestProjects_JiraError(t *testing.T) {
	env := SetupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	resp := env.GET(t, "/projects")

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", resp.StatusCode)
	}
}

func TestProjects_DefaultPagination(t *testing.T) {
	var projects []jiramodels.ProjectResponse

	for i := 0; i < 25; i++ {
		projects = append(projects, makeJiraProject(
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("PROJ%d", i+1),
			fmt.Sprintf("Project %d", i+1),
		))
	}

	env := SetupTestEnv(t, mockJiraHandler(projects, nil))

	resp := env.GET(t, "/projects")

	defer func() {
		_ = resp.Body.Close()
	}()

	var body struct {
		Projects []ProjectItem `json:"projects"`
		PageInfo PageInfo      `json:"pageInfo"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(body.Projects) != 20 {
		t.Errorf("expected 20 projects with default limit, got %d", len(body.Projects))
	}

	if body.PageInfo.PageCount != 2 {
		t.Errorf("expected 2 pages for 25 projects with limit=20, got %d", body.PageInfo.PageCount)
	}
}
