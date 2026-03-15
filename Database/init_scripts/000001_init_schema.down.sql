DROP INDEX IF EXISTS idx_analytics_snapshot_type;
DROP INDEX IF EXISTS idx_analytics_snapshot_project_id;

DROP TABLE IF EXISTS analytics_snapshots;
DROP TABLE IF EXISTS status_changes;
DROP TABLE IF EXISTS issues;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS authors;
DROP USER IF EXISTS pguser;
DROP USER IF EXISTS replicator;
