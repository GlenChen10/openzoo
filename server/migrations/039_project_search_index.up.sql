DO $$
BEGIN
    CREATE INDEX idx_project_title_bigm ON project USING gin (LOWER(title) gin_bigm_ops);
    CREATE INDEX idx_project_description_bigm ON project USING gin (LOWER(COALESCE(description, '')) gin_bigm_ops);
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'pg_bigm not available, skipping project search indexes: %', SQLERRM;
END $$;
