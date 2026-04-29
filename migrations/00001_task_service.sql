-- +goose Up
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS projects(
   id SERIAL PRIMARY KEY,
   name TEXT NOT NULL,
   description TEXT,
   user_id UUID NOT NULL
);


CREATE TABLE IF NOT EXISTS groups(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    project_id INT REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS tasks(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    tags TEXT,
    priority INT NOT NULL,
    status TEXT,
    start_time TIMESTAMP NOT NULL,
    deadline TIMESTAMP,
    group_id INT NOT NULL REFERENCES groups(id),
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
    );
-- +goose Down
SELECT 'down SQL query';

DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS groups CASCADE;
DROP TABLE IF EXISTS projects CASCADE;