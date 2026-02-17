package database

import (
	"database/sql"
	"fmt"
)

var (
	StmtUpsertProject *sql.Stmt
	StmtUpsertIssue   *sql.Stmt
	StmtInsertUser    *sql.Stmt
)

func InitStatements() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	var err error

	StmtUpsertProject, err = db.Prepare(`
		INSERT INTO projects (key, name, url)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE
			SET name = EXCLUDED.name,
			    url  = EXCLUDED.url
		RETURNING id`)
	if err != nil {
		return fmt.Errorf("StmtUpsertProject: %w", err)
	}

	StmtUpsertIssue, err = db.Prepare(`
		INSERT INTO issues (project_id, key, summary, status, created, updated, changelog)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (key) DO UPDATE
			SET summary   = EXCLUDED.summary,
			    status    = EXCLUDED.status,
			    updated   = EXCLUDED.updated,
			    changelog = EXCLUDED.changelog`)
	if err != nil {
		return fmt.Errorf("StmtUpsertIssue: %w", err)
	}

	StmtInsertUser, err = db.Prepare(`
		INSERT INTO users (username, display_name)
		VALUES ($1, $2)
		ON CONFLICT (username) DO NOTHING`)
	if err != nil {
		return fmt.Errorf("StmtInsertUser: %w", err)
	}

	return nil
}

func CloseStatements() {
	for _, s := range []*sql.Stmt{
		StmtUpsertProject,
		StmtUpsertIssue,
		StmtInsertUser,
	} {
		if s != nil {
			s.Close()
		}
	}
}
