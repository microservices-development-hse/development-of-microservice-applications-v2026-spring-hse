INSERT INTO authors (name, external_id, email) VALUES 
('Alice', 'jira-user-101', 'alice@hse.ru'),
('Bob', 'jira-user-102', 'bob@hse.ru'),
('Charlie', 'jira-user-103', 'charlie@hse.ru')
ON CONFLICT (external_id) DO NOTHING;

INSERT INTO projects (key, title) VALUES
('PROJ1', 'Разработка Бекенда'),
('PROJ2', 'Интеграция Коннектора')
ON CONFLICT (key) DO NOTHING;

DO $$
DECLARE
    p1_id INT := (SELECT id FROM projects WHERE key = 'PROJ1' LIMIT 1);
    a1_id INT := (SELECT id FROM authors WHERE external_id = 'jira-user-101' LIMIT 1);
    a2_id INT := (SELECT id FROM authors WHERE external_id = 'jira-user-102' LIMIT 1);
BEGIN
    INSERT INTO issues (
        project_id, author_id, assignee_id, external_id, key, 
        summary, status, priority, type, created_time
    ) VALUES (
        p1_id, a1_id, a2_id, 'ext-task-001', 'PROJ1-1', 
        'Настроить Docker-окружение', 'In Progress', 'High', 'Task', 
        NOW() - INTERVAL '5 days'
    ) ON CONFLICT (external_id) DO NOTHING;

    INSERT INTO issues (
        project_id, author_id, assignee_id, external_id, key, 
        summary, status, priority, type, created_time, closed_time
    ) VALUES (
        p1_id, a2_id, a1_id, 'ext-task-002', 'PROJ1-2', 
        'Исправить баг в миграциях', 'Done', 'Critical', 'Bug', 
        NOW() - INTERVAL '10 days', NOW() - INTERVAL '2 days'
    ) ON CONFLICT (external_id) DO NOTHING;
END $$;

INSERT INTO status_changes (issue_id, author_id, change_time, from_status, to_status)
SELECT 
    id, 
    author_id, 
    now() - interval '4 days', 
    'To Do', 
    'In Progress'
FROM issues 
WHERE key = 'PROJ1-1'
LIMIT 1;

INSERT INTO analytics_snapshots (project_id, type, data)
SELECT 
    id, 
    'velocity', 
    '{"completed_tasks": 1, "open_tasks": 1, "sprint": "Sprint 1"}'::jsonb
FROM projects 
WHERE key = 'PROJ1'
LIMIT 1;