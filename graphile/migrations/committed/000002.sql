--! Previous: sha1:acdabbd00b2371bde1c0370c46c64d55b347997f
--! Hash: sha1:b65a334544d18aed97cedf23c64e74faca31e63c

-- Enter migration here
CREATE TYPE caption_status AS ENUM ('created', 'scheduling', 'running', 'completed');

CREATE TABLE caption_tasks (
    id serial PRIMARY KEY,
    prompt varchar(100) NOT NULL UNIQUE,
    status caption_status NOT NULL DEFAULT 'created',
    total_subtasks int NOT NULL DEFAULT 0,
    completed_subtasks int NOT NULL DEFAULT 0
);
