CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_profiles_email_trgm ON profiles USING gin (email gin_trgm_ops);
