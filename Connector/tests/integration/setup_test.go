package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/microservices-development-hse/connector/internal/database"
	jiraclient "github.com/microservices-development-hse/connector/internal/jira"
	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
	"github.com/microservices-development-hse/connector/internal/server"
)

var globalEnv *TestEnv

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:13",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("pguser"),
		postgres.WithPassword("pgpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to start postgres: %v", err))
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("failed to get connection string: %v", err))
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to open db: %v", err))
	}

	for i := 0; i < 20; i++ {
		if err = db.Ping(); err == nil {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		panic(fmt.Sprintf("db ping failed: %v", err))
	}

	if _, err := db.Exec(schema); err != nil {
		panic(fmt.Sprintf("failed to apply schema: %v", err))
	}

	globalEnv = &TestEnv{
		DB:        db,
		container: pgContainer,
	}

	code := m.Run()

	if err := db.Close(); err != nil {
		fmt.Printf("failed to close db: %v\n", err)
	}

	if err := pgContainer.Terminate(ctx); err != nil {
		fmt.Printf("failed to terminate container: %v\n", err)
	}

	os.Exit(code)
}

// ─── DB schema ──────────────────────────────────────────────────────────────

const schema = `
CREATE TABLE IF NOT EXISTS authors (
    id          SERIAL PRIMARY KEY,
    external_id TEXT UNIQUE,
    name        TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS projects (
    id    SERIAL PRIMARY KEY,
    key   VARCHAR(10) UNIQUE NOT NULL,
    title TEXT NOT NULL,
    url   TEXT
);
CREATE TABLE IF NOT EXISTS issues (
    id           SERIAL PRIMARY KEY,
    external_id  TEXT UNIQUE NOT NULL,
    project_id   INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    author_id    INTEGER REFERENCES authors(id),
    assignee_id  INTEGER REFERENCES authors(id),
    key          TEXT NOT NULL UNIQUE,
    summary      TEXT NOT NULL,
    priority     TEXT,
    status       TEXT,
    created_time TIMESTAMP WITH TIME ZONE,
    closed_time  TIMESTAMP WITH TIME ZONE,
    updated_time TIMESTAMP WITH TIME ZONE,
    time_spent   INTEGER DEFAULT 0
);
CREATE TABLE IF NOT EXISTS status_changes (
    issue_id    INTEGER NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_id   INTEGER REFERENCES authors(id),
    change_time TIMESTAMP WITH TIME ZONE NOT NULL,
    from_status TEXT,
    to_status   TEXT
);
ALTER TABLE status_changes ADD CONSTRAINT status_changes_unique
    UNIQUE (issue_id, author_id, change_time, from_status, to_status);
CREATE TABLE IF NOT EXISTS analytics_snapshots (
    id            SERIAL PRIMARY KEY,
    project_id    INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type          VARCHAR(50) NOT NULL,
    creation_time TIMESTAMP WITH TIME ZONE DEFAULT now(),
    data          JSONB
);
`

// ─── httpClient с таймаутом для всех тестовых запросов ──────────────────────
var testHTTPClient = &http.Client{Timeout: 5 * time.Second}

type TestEnv struct {
	DB         *sql.DB
	JiraMock   *httptest.Server
	HTTPServer *httptest.Server
	container  testcontainers.Container
}

func SetupTestEnv(t *testing.T, jiraHandler http.Handler) *TestEnv {
	t.Helper()

	globalEnv.TruncateTables(t)
	database.ResetForTesting()

	if err := database.SetDBForTesting(globalEnv.DB); err != nil {
		t.Fatalf("failed to inject db: %v", err)
	}

	if err := database.InitStatements(); err != nil {
		t.Fatalf("failed to init statements: %v", err)
	}

	jiraMock := httptest.NewServer(jiraHandler)
	client := jiraclient.NewClient(jiraMock.URL)
	retryConfig := jiraclient.RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 2}
	srv := server.New(0, client, retryConfig, 50, globalEnv.DB, 1)
	connectorSrv := httptest.NewServer(srv.Handler())

	env := &TestEnv{
		DB:         globalEnv.DB,
		JiraMock:   jiraMock,
		HTTPServer: connectorSrv,
		container:  globalEnv.container,
	}

	t.Cleanup(func() {
		database.CloseStatements()
		connectorSrv.Close()
		jiraMock.Close()
		database.ResetForTesting()
	})

	return env
}

func (e *TestEnv) TruncateTables(t *testing.T) {
	t.Helper()

	_, err := e.DB.Exec(`
		TRUNCATE TABLE status_changes, analytics_snapshots, issues, authors, projects
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate failed: %v", err)
	}
}

func (e *TestEnv) CountRows(t *testing.T, table string) int {
	t.Helper()

	var count int

	err := e.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	if err != nil {
		t.Fatalf("count rows in %s failed: %v", table, err)
	}

	return count
}

func (e *TestEnv) GET(t *testing.T, path string) *http.Response {
	t.Helper()

	resp, err := testHTTPClient.Get(e.HTTPServer.URL + path)
	if err != nil {
		t.Fatalf("GET %s failed: %v", path, err)
	}

	return resp
}

func (e *TestEnv) POST(t *testing.T, path string) *http.Response {
	t.Helper()

	resp, err := testHTTPClient.Post(e.HTTPServer.URL+path, "application/json", nil)
	if err != nil {
		t.Fatalf("POST %s failed: %v", path, err)
	}

	return resp
}

// ─── Jira mock helpers ───────────────────────────────────────────────────────
func makeJiraProject(id, key, name string) jiramodels.ProjectResponse {
	return jiramodels.ProjectResponse{
		ID:   id,
		Key:  key,
		Name: name,
		Self: fmt.Sprintf("http://jira/rest/api/2/project/%s", id),
	}
}

func makeJiraIssue(id, key string, withChangelog bool) jiramodels.Issue {
	now := jiramodels.JTime{Time: time.Now()}
	issue := jiramodels.Issue{
		ID:  id,
		Key: key,
		Fields: jiramodels.Fields{
			Summary:      "Summary for " + key,
			Status:       jiramodels.Status{Name: "Open"},
			Priority:     jiramodels.Priority{Name: "High"},
			Creator:      jiramodels.Author{Name: "creator1", DisplayName: "Creator One"},
			Assignee:     &jiramodels.Author{Name: "assignee1", DisplayName: "Assignee One"},
			Created:      now,
			Updated:      now,
			TimeTracking: jiramodels.TimeTracking{TimeSpentSeconds: 3600},
		},
	}

	if withChangelog {
		issue.Changelog = &jiramodels.Changelog{
			Histories: []jiramodels.History{
				{
					Author:  jiramodels.Author{Name: "creator1", DisplayName: "Creator One"},
					Created: now,
					Items:   []jiramodels.Item{{Field: "status", From: "Open", To: "In Progress"}},
				},
			},
		}
	}

	return issue
}

func mockJiraHandler(projects []jiramodels.ProjectResponse, issues []jiramodels.Issue) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/rest/api/2/project":
			_ = json.NewEncoder(w).Encode(projects)

		case "/rest/api/2/search":
			maxResults := r.URL.Query().Get("maxResults")
			if maxResults == "1" {
				total := len(issues)

				var first []jiramodels.Issue

				if total > 0 {
					first = issues[:1]
				}

				_ = json.NewEncoder(w).Encode(jiramodels.IssueSearchResponse{
					Total:  total,
					Issues: first,
				})

				return
			}

			_ = json.NewEncoder(w).Encode(jiramodels.IssueSearchResponse{
				Total:  len(issues),
				Issues: issues,
			})

		default:
			http.NotFound(w, r)
		}
	})
}
