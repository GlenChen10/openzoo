ALTER TABLE agent_task_queue ADD COLUMN IF NOT EXISTS session_id TEXT;
ALTER TABLE agent_task_queue ADD COLUMN IF NOT EXISTS work_dir TEXT;
