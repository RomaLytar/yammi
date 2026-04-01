-- Конвертируем refresh tokens в SHA-256 хэши для безопасного хранения.
-- Существующие сессии будут инвалидированы (старые raw-токены не пройдут hash-lookup).

-- 1. Меняем тип колонки на VARCHAR(64) для SHA-256 hex
ALTER TABLE refresh_tokens ALTER COLUMN token TYPE VARCHAR(64) USING token::text;

-- 2. Хэшируем существующие токены (SHA-256 hex)
UPDATE refresh_tokens SET token = encode(sha256(token::bytea), 'hex');
