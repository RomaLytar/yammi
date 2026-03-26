-- Индекс для фильтрации по типу события (event_type LIKE 'card%')
CREATE INDEX IF NOT EXISTS idx_board_events_board_type ON board_events(board_id, event_type);
