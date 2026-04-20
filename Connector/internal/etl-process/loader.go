package etlprocess

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/microservices-development-hse/connector/internal/logger"
	dbmodels "github.com/microservices-development-hse/connector/internal/models/db"
)

type Loader struct {
	db                     *sql.DB
	stmtUpsertProject      *sql.Stmt
	stmtUpsertIssue        *sql.Stmt
	stmtInsertAuthor       *sql.Stmt
	stmtInsertStatusChange *sql.Stmt
}

type LoaderInterface interface {
	LoadProject(ctx context.Context, project dbmodels.Project) (int, error)
	UpsertAuthors(ctx context.Context, authors map[string]dbmodels.Author) (map[string]int, error)
	LoadIssues(ctx context.Context, issues []dbmodels.Issue) (map[string]int, error)
	LoadStatusChanges(ctx context.Context, changes []dbmodels.StatusChange) error
}

func NewLoader(db *sql.DB, upsertProject, upsertIssue, insertAuthor, insertStatusChange *sql.Stmt) *Loader {
	return &Loader{
		db:                     db,
		stmtUpsertProject:      upsertProject,
		stmtUpsertIssue:        upsertIssue,
		stmtInsertAuthor:       insertAuthor,
		stmtInsertStatusChange: insertStatusChange,
	}
}

func (l *Loader) LoadProject(ctx context.Context, project dbmodels.Project) (int, error) {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("loader: begin tx for project %q failed: %v", project.Key, err)
		return 0, fmt.Errorf("loader: begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	var id int

	err = tx.StmtContext(ctx, l.stmtUpsertProject).QueryRowContext(
		ctx, project.Key, project.Title, project.URL,
	).Scan(&id)
	if err != nil {
		logger.Error("loader: upsert project %q failed: %v", project.Key, err)
		return 0, fmt.Errorf("loader: upsert project %q: %w", project.Key, err)
	}

	if err := tx.Commit(); err != nil {
		logger.Error("loader: commit project %q tx failed: %v", project.Key, err)
		return 0, fmt.Errorf("loader: commit project tx: %w", err)
	}

	logger.Info("loader: project %q loaded with id=%d", project.Key, id)

	return id, nil
}

func (l *Loader) UpsertAuthors(ctx context.Context, authors map[string]dbmodels.Author) (map[string]int, error) {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("loader: begin tx for authors failed: %v", err)
		return nil, fmt.Errorf("loader: begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	stmt := tx.StmtContext(ctx, l.stmtInsertAuthor)
	result := make(map[string]int, len(authors))

	for externalID, author := range authors {
		var id int
		if err := stmt.QueryRowContext(ctx, author.ExternalID, author.Username).Scan(&id); err != nil {
			logger.Error("loader: upsert author %q failed: %v", author.Username, err)
			return nil, fmt.Errorf("loader: upsert author %q: %w", author.Username, err)
		}

		result[externalID] = id
	}

	if err := tx.Commit(); err != nil {
		logger.Error("loader: commit authors tx failed: %v", err)
		return nil, fmt.Errorf("loader: commit authors tx: %w", err)
	}

	logger.Info("loader: successfully upserted %d authors", len(authors))

	return result, nil
}

func (l *Loader) LoadIssues(ctx context.Context, issues []dbmodels.Issue) (map[string]int, error) {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("loader: begin tx for issues failed: %v", err)
		return nil, fmt.Errorf("loader: begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	upsertIssue := tx.StmtContext(ctx, l.stmtUpsertIssue)
	issueIDs := make(map[string]int, len(issues))

	for _, issue := range issues {
		var id int

		err := upsertIssue.QueryRowContext(ctx,
			issue.ExternalID,
			issue.ProjectID,
			issue.AuthorID,
			issue.AssigneeID,
			issue.Key,
			issue.Summary,
			issue.Priority,
			issue.Status,
			issue.CreatedTime,
			issue.UpdatedTime,
			issue.TimeSpent,
		).Scan(&id)
		if err != nil {
			logger.Error("loader: upsert issue %q failed: %v", issue.Key, err)
			return nil, fmt.Errorf("loader: upsert issue %q: %w", issue.Key, err)
		}

		issueIDs[issue.Key] = id
	}

	if err := tx.Commit(); err != nil {
		logger.Error("loader: commit issues tx failed: %v", err)
		return nil, fmt.Errorf("loader: commit issues tx: %w", err)
	}

	logger.Info("loader: successfully loaded %d issues", len(issues))

	return issueIDs, nil
}

func (l *Loader) LoadStatusChanges(ctx context.Context, changes []dbmodels.StatusChange) error {
	if len(changes) == 0 {
		return nil
	}

	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("loader: begin tx for status_changes failed: %v", err)
		return fmt.Errorf("loader: begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	stmt := tx.StmtContext(ctx, l.stmtInsertStatusChange)

	for _, sc := range changes {
		if _, err := stmt.ExecContext(ctx,
			sc.IssueID,
			sc.AuthorID,
			sc.ChangeTime,
			sc.FromStatus,
			sc.ToStatus,
		); err != nil {
			logger.Error("loader: insert status_change for issue_id=%d failed: %v", sc.IssueID, err)
			return fmt.Errorf("loader: insert status_change issue_id=%d: %w", sc.IssueID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error("loader: commit status_changes tx failed: %v", err)
		return fmt.Errorf("loader: commit status_changes tx: %w", err)
	}

	logger.Info("loader: successfully loaded %d status changes", len(changes))

	return nil
}
