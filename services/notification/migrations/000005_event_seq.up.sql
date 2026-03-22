-- Монотонный sequence per board для O(1) unread count (max_seq - last_seen_seq)
-- Заменяет Redis INCR fan-out

-- Добавляем sequence к board_events
ALTER TABLE board_events ADD COLUMN event_seq BIGSERIAL;

-- Индекс для быстрого max(event_seq) per board
CREATE INDEX idx_board_events_board_seq ON board_events(board_id, event_seq DESC);

-- Добавляем last_seen_seq к user_board_cursors
ALTER TABLE user_board_cursors ADD COLUMN last_seen_seq BIGINT NOT NULL DEFAULT 0;
