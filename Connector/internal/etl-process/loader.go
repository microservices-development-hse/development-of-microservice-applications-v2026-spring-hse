package etlprocess

import (
	"context"
	"database/sql"
	"fmt"

	dbmodels "github.com/microservices-development-hse/connector/internal/models/db"
)

type Loader struct {
	db                *sql.DB
	stmtUpsertProject *sql.Stmt
	stmtUpsertIssue   *sql.Stmt
	stmtInsertUser    *sql.Stmt
}

func NewLoader(db *sql.DB, upsertProject, upsertIssue, insertUser *sql.Stmt) *Loader {
	return &Loader{
		db:                db,
		stmtUpsertProject: upsertProject,
		stmtUpsertIssue:   upsertIssue,
		stmtInsertUser:    insertUser,
	}
}

func (l *Loader) LoadProject(ctx context.Context, project dbmodels.Project) (int, error) {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("loader: begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	var id int

	err = tx.StmtContext(ctx, l.stmtUpsertProject).QueryRowContext(
		ctx, project.Key, project.Name, project.URL,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("loader: upsert project %q: %w", project.Key, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("loader: commit project tx: %w", err)
	}

	return id, nil
}

func (l *Loader) LoadIssues(ctx context.Context, issues []dbmodels.Issue, users []dbmodels.User) error {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("loader: begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	upsertIssue := tx.StmtContext(ctx, l.stmtUpsertIssue)
	insertUser := tx.StmtContext(ctx, l.stmtInsertUser)

	for _, u := range users {
		if _, err := insertUser.ExecContext(ctx, u.Username, u.DisplayName); err != nil {
			return fmt.Errorf("loader: insert user %q: %w", u.Username, err)
		}
	}

	for _, issue := range issues {
		if _, err := upsertIssue.ExecContext(ctx,
			issue.ProjectID,
			issue.Key,
			issue.Summary,
			issue.Status,
			issue.Created,
			issue.Updated,
			issue.Changelog,
		); err != nil {
			return fmt.Errorf("loader: upsert issue %q: %w", issue.Key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("loader: commit issues tx: %w", err)
	}

	return nil
}
