--! Previous: sha1:b65a334544d18aed97cedf23c64e74faca31e63c
--! Hash: sha1:feff71f9cffb70177735dcba723077e519a735e5

-- Enter migration here
CREATE OR REPLACE FUNCTION caption_task_status_update_notify()
	RETURNS trigger AS
$$
BEGIN
	PERFORM pg_notify('caption_task_status_channel', '');
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER caption_task_status_update
	AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE ON caption_tasks
	FOR EACH STATEMENT 
EXECUTE PROCEDURE caption_task_status_update_notify();
