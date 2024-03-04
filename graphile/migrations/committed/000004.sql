--! Previous: sha1:feff71f9cffb70177735dcba723077e519a735e5
--! Hash: sha1:09b35c2ffa54a319319c020f574278d07dee869d

-- Enter migration here

ALTER TABLE caption_tasks ADD COLUMN created_at timestamptz NOT NULL DEFAULT now();
