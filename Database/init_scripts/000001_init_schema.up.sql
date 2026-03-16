DO $$ 
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'pguser') THEN
    CREATE USER pguser WITH PASSWORD 'pgpassword';
  END IF;
END $$;

GRANT ALL PRIVILEGES ON DATABASE testdb TO pguser;
ALTER DATABASE testdb OWNER TO pguser;

GRANT ALL ON SCHEMA public TO pguser;


CREATE TABLE authors (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(255) UNIQUE NOT NULL,
    name TEXT NOT NULL,
    email TEXT
);

CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    key VARCHAR(10) UNIQUE NOT NULL,
    title TEXT NOT NULL
);

CREATE TABLE issues (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(255) UNIQUE NOT NULL,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    author_id INTEGER REFERENCES authors(id),
    assignee_id INTEGER REFERENCES authors(id),
    key TEXT NOT NULL UNIQUE,
    summary TEXT NOT NULL,
    description TEXT,
    type TEXT,
    priority TEXT,
    status TEXT,
    created_time TIMESTAMP WITH TIME ZONE,
    closed_time TIMESTAMP WITH TIME ZONE,
    updated_time TIMESTAMP WITH TIME ZONE,
    time_spent INTEGER DEFAULT 0
);

CREATE TABLE status_changes (
    issue_id INTEGER NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_id INTEGER NOT NULL REFERENCES authors(id),
    change_time TIMESTAMP WITH TIME ZONE NOT NULL,
    from_status TEXT,
    to_status TEXT
);

CREATE TABLE analytics_snapshots (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    creation_time TIMESTAMP WITH TIME ZONE DEFAULT now(),
    data JSONB
);

CREATE INDEX "idx_analytics_snapshot_project_id" ON analytics_snapshots ("project_id");
CREATE INDEX "idx_analytics_snapshot_type" ON analytics_snapshots ("type");

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO pguser;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO pguser;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO pguser;
