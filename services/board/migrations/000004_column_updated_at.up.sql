ALTER TABLE columns ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
UPDATE columns SET updated_at = created_at;
