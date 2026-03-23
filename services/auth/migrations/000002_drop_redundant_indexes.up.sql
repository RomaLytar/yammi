-- idx_users_email дублирует индекс от UNIQUE constraint на users.email
DROP INDEX IF EXISTS idx_users_email;

-- idx_refresh_tokens_token дублирует индекс от UNIQUE constraint на refresh_tokens.token
DROP INDEX IF EXISTS idx_refresh_tokens_token;
