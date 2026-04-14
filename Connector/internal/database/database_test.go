package database

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/microservices-development-hse/connector/config"
)

// ─── helpers ────────────────────────────────────────────────────────────────

func setupMockDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()

	conn, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	db = conn

	t.Cleanup(func() {
		db = nil
	})

	return mock
}

// ─── Init ───────────────────────────────────────────────────────────────────

func TestInit_AlreadyInitialized(t *testing.T) {
	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	db = conn

	defer func() { db = nil }()

	err = Init(config.DBSettings{})
	if err == nil {
		t.Fatal("expected error when db already initialized")
	}
}

func TestInit_InvalidConnectionString(t *testing.T) {
	cfg := config.DBSettings{
		User:     "invalid",
		Password: "invalid",
		Host:     "localhost",
		Port:     9999,
		Name:     "nonexistent",
	}

	err := Init(cfg)
	if err == nil {
		t.Fatal("expected error for invalid connection")
	}

	db = nil
}

// ─── GetDB ──────────────────────────────────────────────────────────────────

func TestGetDB_PanicsWhenNotInitialized(t *testing.T) {
	db = nil

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when db not initialized")
		}
	}()

	GetDB()
}

func TestGetDB_ReturnsDB(t *testing.T) {
	setupMockDB(t)

	result := GetDB()
	if result == nil {
		t.Fatal("expected non-nil db")
	}
}

// ─── Close ──────────────────────────────────────────────────────────────────

func TestClose_WhenNil(t *testing.T) {
	db = nil

	err := Close()
	if err != nil {
		t.Errorf("expected no error when db is nil, got %v", err)
	}
}

func TestClose_Success(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectClose()

	err := Close()
	if err != nil {
		t.Errorf("expected no error on close, got %v", err)
	}

	if db != nil {
		t.Error("expected db to be nil after close")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestClose_SetsNilOnError(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectClose().WillReturnError(fmt.Errorf("close error"))

	err := Close()
	if err == nil {
		t.Fatal("expected error on close failure")
	}

	if db != nil {
		t.Error("expected db to be nil even after close error")
	}
}

// ─── Ping ───────────────────────────────────────────────────────────────────

func TestPing_WhenNotInitialized(t *testing.T) {
	db = nil

	err := Ping()
	if err == nil {
		t.Fatal("expected error when db not initialized")
	}
}

func TestPing_Success(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectPing()

	err := Ping()
	if err != nil {
		t.Errorf("expected no error on ping, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPing_Failure(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectPing().WillReturnError(fmt.Errorf("ping error"))

	err := Ping()
	if err == nil {
		t.Fatal("expected error on ping failure")
	}
}

// ─── InitStatements ─────────────────────────────────────────────────────────

func TestInitStatements_WhenDBNil(t *testing.T) {
	db = nil

	err := InitStatements()
	if err == nil {
		t.Fatal("expected error when db not initialized")
	}
}

func TestInitStatements_Success(t *testing.T) {
	mock := setupMockDB(t)

	mock.ExpectPrepare("INSERT INTO projects")
	mock.ExpectPrepare("INSERT INTO issues")
	mock.ExpectPrepare("INSERT INTO authors")
	mock.ExpectPrepare("INSERT INTO status_changes")

	err := InitStatements()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}

	CloseStatements()
}

func TestInitStatements_FailOnProject(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectPrepare("INSERT INTO projects").WillReturnError(fmt.Errorf("prepare error"))

	err := InitStatements()
	if err == nil {
		t.Fatal("expected error when prepare project fails")
	}
}

func TestInitStatements_FailOnIssue(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectPrepare("INSERT INTO projects")
	mock.ExpectPrepare("INSERT INTO issues").WillReturnError(fmt.Errorf("prepare error"))

	err := InitStatements()
	if err == nil {
		t.Fatal("expected error when prepare issue fails")
	}

	CloseStatements()
}

func TestInitStatements_FailOnAuthor(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectPrepare("INSERT INTO projects")
	mock.ExpectPrepare("INSERT INTO issues")
	mock.ExpectPrepare("INSERT INTO authors").WillReturnError(fmt.Errorf("prepare error"))

	err := InitStatements()
	if err == nil {
		t.Fatal("expected error when prepare author fails")
	}

	CloseStatements()
}

func TestInitStatements_FailOnStatusChange(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectPrepare("INSERT INTO projects")
	mock.ExpectPrepare("INSERT INTO issues")
	mock.ExpectPrepare("INSERT INTO authors")
	mock.ExpectPrepare("INSERT INTO status_changes").WillReturnError(fmt.Errorf("prepare error"))

	err := InitStatements()
	if err == nil {
		t.Fatal("expected error when prepare status_change fails")
	}

	CloseStatements()
}

// ─── CloseStatements ────────────────────────────────────────────────────────

func TestCloseStatements_NilStatements(t *testing.T) {
	StmtUpsertProject = nil
	StmtUpsertIssue = nil
	StmtInsertAuthor = nil
	StmtInsertStatusChange = nil

	CloseStatements()
}

func TestClose_ErrorOnClose(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectClose().WillReturnError(fmt.Errorf("close error"))

	err := Close()
	if err == nil {
		t.Fatal("expected error on close failure")
	}

	if db != nil {
		t.Error("expected db to be nil even after close error")
	}
}

func TestCloseStatements_WithClosedStmt(t *testing.T) {
	mock := setupMockDB(t)
	mock.ExpectPrepare("INSERT INTO projects")
	mock.ExpectPrepare("INSERT INTO issues")
	mock.ExpectPrepare("INSERT INTO authors")
	mock.ExpectPrepare("INSERT INTO status_changes")

	_ = InitStatements()

	mock.ExpectClose()

	_ = db.Close()

	CloseStatements()
}
