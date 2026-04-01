-- Невозможно восстановить оригинальные токены из хэшей.
-- Просто меняем тип обратно и ревокаем все токены.
UPDATE refresh_tokens SET revoked = TRUE;
ALTER TABLE refresh_tokens ALTER COLUMN token TYPE UUID USING gen_random_uuid();
