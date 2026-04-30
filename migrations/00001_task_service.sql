-- +goose Up
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS projects(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS groups(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    project_id INT REFERENCES projects(id) ON DELETE CASCADE,
);

CREATE TABLE IF NOT EXISTS tasks(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    tags TEXT,
    priority INT NOT NULL,
    status TEXT,
    start_time TIMESTAMPTZ NOT NULL,
    deadline TIMESTAMPTZ,
    group_id INT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_archived BOOLEAN DEFAULT false
    );

CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_groups_project_id ON groups(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_group_id ON tasks(group_id);
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
-- +goose Down
SELECT 'down SQL query';

DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS groups CASCADE;
DROP TABLE IF EXISTS projects CASCADE;