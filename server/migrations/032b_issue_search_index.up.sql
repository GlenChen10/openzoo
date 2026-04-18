DO $$
BEGIN
    CREATE EXTENSION IF NOT EXISTS pg_bigm;
    CREATE INDEX idx_issue_title_bigm ON issue USING gin (title gin_bigm_ops);
    CREATE INDEX idx_issue_description_bigm ON issue USING gin (COALESCE(description, '') gin_bigm_ops);
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'pg_bigm not available, skipping search indexes: %', SQLERRM;
END $$;
