package database

import (
	"database/sql"
	"fmt"

	"github.com/microservices-development-hse/connector/internal/logger"
)

var (
	StmtUpsertProject      *sql.Stmt
	StmtUpsertIssue        *sql.Stmt
	StmtInsertAuthor       *sql.Stmt
	StmtInsertStatusChange *sql.Stmt
)

func InitStatements() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	var err error

	StmtUpsertProject, err = db.Prepare(`
		INSERT INTO projects (key, title, url)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE
			SET title = EXCLUDED.title,
			    url   = EXCLUDED.url
		RETURNING id`)
	if err != nil {
		return fmt.Errorf("StmtUpsertProject: %w", err)
	}

	StmtUpsertIssue, err = db.Prepare(`
		INSERT INTO issues (external_id, project_id, author_id, assignee_id, key, summary, priority, status, created_time, updated_time, time_spent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (key) DO UPDATE
			SET summary      = EXCLUDED.summary,
			    status       = EXCLUDED.status,
			    priority     = EXCLUDED.priority,
			    assignee_id  = EXCLUDED.assignee_id,
			    updated_time = EXCLUDED.updated_time,
			    time_spent   = EXCLUDED.time_spent
		RETURNING id`)
	if err != nil {
		return fmt.Errorf("StmtUpsertIssue: %w", err)
	}

	StmtInsertAuthor, err = db.Prepare(`
		INSERT INTO authors (external_id, name)
		VALUES ($1, $2)
		ON CONFLICT (external_id) DO UPDATE
			SET name = EXCLUDED.name
		RETURNING id`)
	if err != nil {
		return fmt.Errorf("StmtInsertAuthor: %w", err)
	}

	StmtInsertStatusChange, err = db.Prepare(`
		INSERT INTO status_changes (issue_id, author_id, change_time, from_status, to_status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (issue_id, author_id, change_time, from_status, to_status) DO NOTHING`)
	if err != nil {
		return fmt.Errorf("StmtInsertStatusChange: %w", err)
	}

	return nil
}

func CloseStatements() {
	for _, s := range []*sql.Stmt{
		StmtUpsertProject,
		StmtUpsertIssue,
		StmtInsertAuthor,
		StmtInsertStatusChange,
	} {
		if s != nil {
			if err := s.Close(); err != nil {
				logger.Error("failed to close statement: %v", err)
			}
		}
	}
}
