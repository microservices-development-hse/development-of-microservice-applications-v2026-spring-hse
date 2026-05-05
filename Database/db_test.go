package postgres

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

// These tests are integration tests for the PostgreSQL database used by the project.
//
// Expected environment:
//   - PostgreSQL is running (e.g. via Database/docker-compose.yml)
//   - The database is reachable at DATABASE_TEST_DSN
//
// Example DSN:
//   host=localhost port=5432 user=pguser password=pgpassword dbname=testdb sslmode=disable

func testDSN(t *testing.T) string {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("DATABASE_TEST_DSN"))
	if dsn == "" {
		dsn = "host=localhost port=5432 user=pguser password=pgpassword dbname=testdb sslmode=disable"
	}
	return dsn
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", testDSN(t))
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	return db
}

func assertTableExists(t *testing.T, db *sql.DB, table string) {
	t.Helper()

	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = $1
		)
	`, table).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check table %q: %v", table, err)
	}

	if !exists {
		t.Fatalf("expected table %q to exist", table)
	}
}

func assertColumnsExist(t *testing.T, db *sql.DB, table string, columns ...string) {
	t.Helper()

	for _, column := range columns {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = 'public'
				  AND table_name = $1
				  AND column_name = $2
			)
		`, table, column).Scan(&exists)
		if err != nil {
			t.Fatalf("failed to check column %q.%q: %v", table, column, err)
		}

		if !exists {
			t.Fatalf("expected column %q to exist in table %q", column, table)
		}
	}
}

func TestDatabaseSchema(t *testing.T) {
	db := openTestDB(t)

	for _, table := range []string{
		"authors",
		"projects",
		"issues",
		"status_changes",
		"analytics_snapshots",
	} {
		assertTableExists(t, db, table)
	}

	assertColumnsExist(t, db, "projects", "id", "key", "title", "url")
	assertColumnsExist(t, db, "authors", "id", "external_id", "name")
	assertColumnsExist(t, db, "issues",
		"id", "external_id", "project_id", "author_id", "assignee_id",
		"key", "summary", "priority", "status", "created_time",
		"closed_time", "updated_time", "time_spent",
	)
	assertColumnsExist(t, db, "status_changes",
		"issue_id", "author_id", "change_time", "from_status", "to_status",
	)
	assertColumnsExist(t, db, "analytics_snapshots",
		"id", "project_id", "type", "creation_time", "data",
	)
}

func TestDatabaseSeedData(t *testing.T) {
	db := openTestDB(t)

	var projectsCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM projects`).Scan(&projectsCount); err != nil {
		t.Fatalf("failed to count projects: %v", err)
	}
	if projectsCount < 2 {
		t.Fatalf("expected at least 2 projects, got %d", projectsCount)
	}

	for _, key := range []string{"PROJ1", "PROJ2"} {
		var exists bool
		err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM projects WHERE key = $1)`, key).Scan(&exists)
		if err != nil {
			t.Fatalf("failed to check project %q: %v", key, err)
		}
		if !exists {
			t.Fatalf("expected project %q to exist", key)
		}
	}

	var issuesCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM issues`).Scan(&issuesCount); err != nil {
		t.Fatalf("failed to count issues: %v", err)
	}
	if issuesCount < 3 {
		t.Fatalf("expected at least 3 issues, got %d", issuesCount)
	}

	for _, issueKey := range []string{"PROJ1-1", "PROJ1-2", "PROJ2-1"} {
		var exists bool
		err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM issues WHERE key = $1)`, issueKey).Scan(&exists)
		if err != nil {
			t.Fatalf("failed to check issue %q: %v", issueKey, err)
		}
		if !exists {
			t.Fatalf("expected issue %q to exist", issueKey)
		}
	}

	var snapshotsCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM analytics_snapshots`).Scan(&snapshotsCount); err != nil {
		t.Fatalf("failed to count analytics snapshots: %v", err)
	}
	if snapshotsCount < 2 {
		t.Fatalf("expected at least 2 analytics snapshots, got %d", snapshotsCount)
	}

	for _, snapshotType := range []string{"velocity", "complexity"} {
		var exists bool
		err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM analytics_snapshots WHERE type = $1)`, snapshotType).Scan(&exists)
		if err != nil {
			t.Fatalf("failed to check analytics snapshot type %q: %v", snapshotType, err)
		}
		if !exists {
			t.Fatalf("expected analytics snapshot type %q to exist", snapshotType)
		}
	}
}
