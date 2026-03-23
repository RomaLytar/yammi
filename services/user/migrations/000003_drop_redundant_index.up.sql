-- idx_profiles_email дублирует индекс от UNIQUE constraint на profiles.email
-- (тригам-индекс idx_profiles_email_trgm остаётся — он для ILIKE поиска)
DROP INDEX IF EXISTS idx_profiles_email;
