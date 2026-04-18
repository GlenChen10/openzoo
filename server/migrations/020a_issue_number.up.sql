ALTER TABLE workspace
    ADD COLUMN IF NOT EXISTS issue_prefix TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS issue_counter INT NOT NULL DEFAULT 0;
ALTER TABLE issue ADD COLUMN IF NOT EXISTS number INT NOT NULL DEFAULT 0;

UPDATE workspace SET issue_prefix = UPPER(
    LEFT(REGEXP_REPLACE(name, '[^a-zA-Z]', '', 'g'), 3)
);
UPDATE workspace SET issue_prefix = 'OZ' WHERE issue_prefix = '';

WITH numbered AS (
    SELECT id, workspace_id,
           ROW_NUMBER() OVER (PARTITION BY workspace_id ORDER BY created_at ASC) AS rn
    FROM issue
)
UPDATE issue SET number = numbered.rn FROM numbered WHERE issue.id = numbered.id;

UPDATE workspace SET issue_counter = COALESCE(
    (SELECT MAX(number) FROM issue WHERE issue.workspace_id = workspace.id), 0
);

ALTER TABLE issue ADD CONSTRAINT uq_issue_workspace_number UNIQUE (workspace_id, number);
CREATE INDEX idx_issue_workspace_number ON issue(workspace_id, number);
