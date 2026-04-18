DROP INDEX IF EXISTS idx_issue_title_bigm;
DROP INDEX IF EXISTS idx_issue_description_bigm;
DROP INDEX IF EXISTS idx_comment_content_bigm;

DO $$
BEGIN
    CREATE INDEX idx_issue_title_bigm ON issue USING gin (LOWER(title) gin_bigm_ops);
    CREATE INDEX idx_issue_description_bigm ON issue USING gin (LOWER(COALESCE(description, '')) gin_bigm_ops);
    CREATE INDEX idx_comment_content_bigm ON comment USING gin (LOWER(content) gin_bigm_ops);
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'pg_bigm not available, skipping search indexes: %', SQLERRM;
END $$;
