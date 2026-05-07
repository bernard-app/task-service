-- +goose Up
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS projects(
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS groups(
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    project_id BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks(
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    priority INT NOT NULL,
    status TEXT,
    start_time TIMESTAMPTZ NOT NULL,
    deadline TIMESTAMPTZ,
    group_id BIGINT REFERENCES groups(id) ON DELETE CASCADE,
    project_id BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_archived BOOLEAN DEFAULT false
);

CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    color VARCHAR(7) NOT NULL,
    is_system BOOLEAN DEFAULT false,
    user_id UUID NOT NULL,
    project_id BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks_tags (
    task_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE,
    tag_id BIGINT REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY(task_id, tag_id)
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