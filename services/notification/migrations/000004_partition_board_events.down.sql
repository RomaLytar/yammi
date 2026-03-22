-- Откат: обратно в обычную таблицу
CREATE TABLE board_events_flat (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL,
    actor_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO board_events_flat SELECT * FROM board_events;
DROP TABLE board_events;
ALTER TABLE board_events_flat RENAME TO board_events;
CREATE INDEX idx_board_events_board_created ON board_events(board_id, created_at DESC);
