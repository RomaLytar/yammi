-- Добавляем денормализованное поле reply_count
ALTER TABLE comments ADD COLUMN reply_count INTEGER NOT NULL DEFAULT 0;

-- Заполняем существующие значения
UPDATE comments SET reply_count = (
    SELECT COUNT(*) FROM comments c2 WHERE c2.parent_id = comments.id
);
