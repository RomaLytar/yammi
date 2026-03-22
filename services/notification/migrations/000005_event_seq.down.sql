ALTER TABLE user_board_cursors DROP COLUMN last_seen_seq;
DROP INDEX IF EXISTS idx_board_events_board_seq;
ALTER TABLE board_events DROP COLUMN event_seq;
