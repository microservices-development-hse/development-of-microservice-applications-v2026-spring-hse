CREATE TABLE IF NOT EXISTS authors (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
  id SERIAL PRIMARY KEY,
  key TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS issues (
  id SERIAL PRIMARY KEY,
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  author_id INTEGER REFERENCES authors(id),
  assignee_id INTEGER REFERENCES authors(id),
  key TEXT NOT NULL,
  summary TEXT,
  description TEXT,
  type TEXT,
  priority TEXT,
  status TEXT,
  created_time TIMESTAMP WITH TIME ZONE DEFAULT now(),
  closed_time TIMESTAMP WITH TIME ZONE,
  updated_time TIMESTAMP WITH TIME ZONE,
  time_spent INTEGER
);

CREATE INDEX IF NOT EXISTS idx_issues_project_id ON issues(project_id);
CREATE INDEX IF NOT EXISTS idx_issues_key ON issues(key);
CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);

CREATE TABLE IF NOT EXISTS status_changes (
  id SERIAL PRIMARY KEY,
  issue_id INTEGER NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  author_id INTEGER REFERENCES authors(id),
  change_time TIMESTAMP WITH TIME ZONE DEFAULT now(),
  from_status TEXT,
  to_status TEXT
);

CREATE INDEX IF NOT EXISTS idx_status_changes_issue_id ON status_changes(issue_id);
CREATE INDEX IF NOT EXISTS idx_status_changes_change_time ON status_changes(change_time);

CREATE TABLE IF NOT EXISTS open_task_time (
  id SERIAL PRIMARY KEY,
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  creation_time TIMESTAMP WITH TIME ZONE DEFAULT now(),
  data JSONB
);

CREATE TABLE IF NOT EXISTS task_state_time (
  id SERIAL PRIMARY KEY,
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  creation_time TIMESTAMP WITH TIME ZONE DEFAULT now(),
  state TEXT,
  data JSONB
);

CREATE TABLE IF NOT EXISTS complexity_task_time (
  id SERIAL PRIMARY KEY,
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  creation_time TIMESTAMP WITH TIME ZONE DEFAULT now(),
  data JSONB
);

CREATE TABLE IF NOT EXISTS task_priority_count (
  id SERIAL PRIMARY KEY,
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  creation_time TIMESTAMP WITH TIME ZONE DEFAULT now(),
  state TEXT,
  data JSONB
);

CREATE TABLE IF NOT EXISTS activity_by_task (
  id SERIAL PRIMARY KEY,
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  creation_time TIMESTAMP WITH TIME ZONE DEFAULT now(),
  state TEXT,
  data JSONB
);

CREATE INDEX IF NOT EXISTS idx_analytics_project_creation ON open_task_time(project_id, creation_time);
CREATE INDEX IF NOT EXISTS idx_task_state_project_creation ON task_state_time(project_id, creation_time);

CREATE TABLE IF NOT EXISTS graphs (
  id SERIAL PRIMARY KEY,
  project TEXT NOT NULL,
  task TEXT NOT NULL,
  data JSONB,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_graphs_project_task ON graphs(project, task);
