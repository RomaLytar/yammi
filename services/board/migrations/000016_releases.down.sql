ALTER TABLE board_settings DROP COLUMN IF EXISTS done_column_id;

DROP INDEX IF EXISTS idx_cards_backlog;
DROP INDEX IF EXISTS idx_cards_release_id;
ALTER TABLE cards DROP COLUMN IF EXISTS release_id;

DROP TABLE IF EXISTS releases;
