INSERT INTO authors (name) VALUES ('Alice') ON CONFLICT DO NOTHING;
INSERT INTO authors (name) VALUES ('Bob') ON CONFLICT DO NOTHING;

INSERT INTO projects (key, title) VALUES
('PROJ1', 'Project One') ON CONFLICT (key) DO NOTHING,
('PROJ2', 'Project Two') ON CONFLICT (key) DO NOTHING;

WITH p AS (SELECT id FROM projects WHERE key = 'PROJ1' LIMIT 1),
     a AS (SELECT id FROM authors LIMIT 1)
INSERT INTO issues (project_id, author_id, assignee_id, key, summary, description, type, priority, status, created_time)
SELECT p.id, (SELECT id FROM authors LIMIT 1), (SELECT id FROM authors LIMIT 1), 'PROJ1-1', 'First issue', 'Description', 'Bug', 'Major', 'In Progress', now() - interval '5 days' FROM p
ON CONFLICT DO NOTHING;

WITH p AS (SELECT id FROM projects WHERE key = 'PROJ1' LIMIT 1),
     a AS (SELECT id FROM authors LIMIT 1)
INSERT INTO issues (project_id, author_id, assignee_id, key, summary, description, type, priority, status, created_time, closed_time)
SELECT p.id, (SELECT id FROM authors LIMIT 1), (SELECT id FROM authors LIMIT 1), 'PROJ1-2', 'Done issue', 'Desc', 'Task', 'Minor', 'Done', now() - interval '15 days', now() - interval '2 days' FROM p
ON CONFLICT DO NOTHING;

WITH i AS (SELECT id FROM issues WHERE key='PROJ1-1' LIMIT 1)
INSERT INTO status_changes (issue_id, author_id, change_time, from_status, to_status)
SELECT i.id, (SELECT id FROM authors LIMIT 1), now() - interval '4 days', 'To Do', 'In Progress' FROM i
ON CONFLICT DO NOTHING;

INSERT INTO graphs (project, task, data) VALUES ('PROJ1', '1', '[]'::jsonb) ON CONFLICT DO NOTHING;
INSERT INTO open_task_time (project_id, data) SELECT id, '[]'::jsonb FROM projects WHERE key='PROJ1' LIMIT 1 ON CONFLICT DO NOTHING;
