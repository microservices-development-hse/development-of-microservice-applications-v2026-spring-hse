INSERT INTO authors (name, external_id) VALUES 
('Alice', 'ext-auth-101'),
('Bob', 'ext-auth-102'),
('Charlie', 'ext-auth-103')
ON CONFLICT (external_id) DO NOTHING;

INSERT INTO projects (key, title) VALUES
('PROJ1', 'Разработка Бекенда'),
('PROJ2', 'Интеграция Коннектора')
ON CONFLICT (key) DO NOTHING;

DO $$
DECLARE 
    p1_id INT := (SELECT id FROM projects WHERE key = 'PROJ1' LIMIT 1);
    a1_id INT := (SELECT id FROM authors WHERE name = 'Alice' LIMIT 1);
    a2_id INT := (SELECT id FROM authors WHERE name = 'Bob' LIMIT 1);
BEGIN
    INSERT INTO issues (
        project_id, author_id, assignee_id, external_id, key, 
        summary, status, priority, created_time
    ) VALUES (
        p1_id, a1_id, a2_id, 'jira-task-001', 'PROJ1-1', 
        'Настроить Docker-окружение', 'In Progress', 'High', 
        NOW() - INTERVAL '5 days'
    ) ON CONFLICT (external_id) DO NOTHING;

    INSERT INTO issues (
        project_id, author_id, assignee_id, external_id, key, 
        summary, status, priority, created_time, closed_time
    ) VALUES (
        p1_id, a2_id, a1_id, 'jira-task-002', 'PROJ1-2', 
        'Исправить баг в миграциях', 'Done', 'Critical', 
        NOW() - INTERVAL '10 days', NOW() - INTERVAL '2 days'
    ) ON CONFLICT (external_id) DO NOTHING;

    INSERT INTO issues (
        project_id, author_id, assignee_id, external_id, key,
        summary, status, priority, created_time
    ) VALUES (
        p1_id, a1_id, a2_id, 'jira-task-003', 'PROJ1-3',
        'Тест переоткрытия', 'Reopened', 'Medium',NOW() - INTERVAL '1 day'
    ) ON CONFLICT (external_id) DO NOTHING;

INSERT INTO issues (project_id, author_id, assignee_id, external_id, key, summary, status, priority, created_time)
VALUES (p1_id, a1_id, a2_id, 'jira-task-004', 'PROJ1-4', 'Тест резолва', 'Resolved', 'Low', NOW() - INTERVAL '1 day');
END $$;

DO $$ 
DECLARE 
    p2_id INT := (SELECT id FROM projects WHERE key = 'PROJ2' LIMIT 1);
    a1_id INT := (SELECT id FROM authors WHERE name = 'Alice' LIMIT 1);
BEGIN 
    INSERT INTO issues (
        project_id, author_id, assignee_id, external_id, key, 
        summary, status, priority, created_time, closed_time
    ) VALUES (
        p2_id, a1_id, a1_id, 'jira-task-999', 'PROJ2-1', 
        'Тестовая закрытая задача', 'Done', 'Low', 
        NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 hour'
    ) ON CONFLICT (external_id) DO NOTHING;

    INSERT INTO status_changes (issue_id, author_id, change_time, from_status, to_status)
    SELECT id, a1_id, NOW() - INTERVAL '1 day', 'To Do', 'In Progress' 
    FROM issues WHERE key = 'PROJ2-1';
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

INSERT INTO status_changes (issue_id, author_id, change_time, from_status, to_status)
SELECT 
    id, 
    author_id, 
    now() - interval '2 days', 
    'In Progress', 
    'Testing'
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

INSERT INTO analytics_snapshots (project_id, type, data)
SELECT 
    id, 
    'complexity', 
    '[{"issue_key": "PROJ2-1", "lead_time": 47, "move_count": 1}]'::jsonb
FROM projects 
WHERE key = 'PROJ2'
LIMIT 1;