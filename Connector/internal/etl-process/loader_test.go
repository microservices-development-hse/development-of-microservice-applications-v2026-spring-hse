package etlprocess

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	dbmodels "github.com/microservices-development-hse/connector/internal/models/db"
)

func setup(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Loader) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	mock.ExpectPrepare("INSERT INTO project")

	stmtProject, _ := db.Prepare("INSERT INTO project")

	mock.ExpectPrepare("INSERT INTO issue")

	stmtIssue, _ := db.Prepare("INSERT INTO issue")

	mock.ExpectPrepare("INSERT INTO author")

	stmtAuthor, _ := db.Prepare("INSERT INTO author")

	mock.ExpectPrepare("INSERT INTO status")

	stmtStatus, _ := db.Prepare("INSERT INTO status")

	loader := NewLoader(db, stmtProject, stmtIssue, stmtAuthor, stmtStatus)

	return db, mock, loader
}

func TestNewLoader(t *testing.T) {
	db, _, l := setup(t)

	if l == nil {
		t.Fatal("loader is nil")
	}

	if l.db != db {
		t.Error("db not set")
	}
}

func TestLoadProject_Success(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	project := dbmodels.Project{
		Key:   "TEST",
		Title: "Test",
		URL:   "url",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO project")).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	id, err := l.LoadProject(context.Background(), project)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 1 {
		t.Errorf("expected id=1, got %d", id)
	}
}

func TestLoadProject_BeginError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	_, err := l.LoadProject(context.Background(), dbmodels.Project{Key: "X"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertAuthors_Success(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	authors := map[string]dbmodels.Author{
		"a": {ExternalID: "a", Username: "A"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO author").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectCommit()

	res, err := l.UpsertAuthors(context.Background(), authors)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res["a"] != 10 {
		t.Errorf("expected id=10, got %d", res["a"])
	}
}

func TestLoadIssues_Success(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	now := time.Now()

	issues := []dbmodels.Issue{
		{
			ExternalID:  "1",
			ProjectID:   1,
			Key:         "TEST-1",
			Summary:     "test",
			Priority:    "High",
			Status:      "Open",
			CreatedTime: now,
			UpdatedTime: now,
		},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO issue").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))
	mock.ExpectCommit()

	res, err := l.LoadIssues(context.Background(), issues)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res["TEST-1"] != 100 {
		t.Errorf("expected id=100, got %d", res["TEST-1"])
	}
}

func TestLoadStatusChanges_Empty(t *testing.T) {
	db, _, l := setup(t)

	defer func() { _ = db.Close() }()

	err := l.LoadStatusChanges(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestLoadStatusChanges_Success(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	changes := []dbmodels.StatusChange{
		{
			IssueID:    1,
			AuthorID:   1,
			ChangeTime: time.Now(),
			FromStatus: "Open",
			ToStatus:   "Closed",
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO status").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := l.LoadStatusChanges(context.Background(), changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadProject_QueryError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO project").WillReturnError(fmt.Errorf("query error"))

	_, err := l.LoadProject(context.Background(), dbmodels.Project{Key: "TEST"})
	if err == nil {
		t.Fatal("expected error")
	}
}
func TestLoadProject_CommitError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO project").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	_, err := l.LoadProject(context.Background(), dbmodels.Project{Key: "TEST"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertAuthors_QueryError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	authors := map[string]dbmodels.Author{
		"a": {ExternalID: "a", Username: "A"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO author").WillReturnError(fmt.Errorf("query error"))

	_, err := l.UpsertAuthors(context.Background(), authors)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertAuthors_CommitError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	authors := map[string]dbmodels.Author{
		"a": {ExternalID: "a", Username: "A"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO author").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	_, err := l.UpsertAuthors(context.Background(), authors)
	if err == nil {
		t.Fatal("expected error")
	}
}
func TestLoadIssues_QueryError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	issues := []dbmodels.Issue{
		{Key: "TEST-1"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO issue").WillReturnError(fmt.Errorf("query error"))

	_, err := l.LoadIssues(context.Background(), issues)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadIssues_CommitError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	issues := []dbmodels.Issue{
		{Key: "TEST-1"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO issue").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	_, err := l.LoadIssues(context.Background(), issues)
	if err == nil {
		t.Fatal("expected error")
	}
}
func TestLoadStatusChanges_ExecError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	changes := []dbmodels.StatusChange{
		{IssueID: 1},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO status").WillReturnError(fmt.Errorf("exec error"))

	err := l.LoadStatusChanges(context.Background(), changes)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadStatusChanges_CommitError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	changes := []dbmodels.StatusChange{
		{IssueID: 1},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO status").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	err := l.LoadStatusChanges(context.Background(), changes)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertAuthors_BeginError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	authors := map[string]dbmodels.Author{
		"a": {ExternalID: "a", Username: "A"},
	}

	_, err := l.UpsertAuthors(context.Background(), authors)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadIssues_BeginError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	issues := []dbmodels.Issue{
		{Key: "TEST-1"},
	}

	_, err := l.LoadIssues(context.Background(), issues)
	if err == nil {
		t.Fatal("expected error")
	}
}
func TestLoadStatusChanges_BeginError(t *testing.T) {
	db, mock, l := setup(t)

	defer func() { _ = db.Close() }()

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	changes := []dbmodels.StatusChange{
		{IssueID: 1},
	}

	err := l.LoadStatusChanges(context.Background(), changes)
	if err == nil {
		t.Fatal("expected error")
	}
}
