-- Убираем partitioning board_events — оно замедляет MAX/JOIN queries.
-- Одна таблица с индексом быстрее для нашего workload (read-heavy).

CREATE TABLE board_events_flat (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL,
    actor_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event_seq BIGSERIAL,
    PRIMARY KEY (id)
);

INSERT INTO board_events_flat (id, board_id, actor_id, event_type, title, message, metadata, created_at, event_seq)
SELECT id, board_id, actor_id, event_type, title, message, metadata, created_at, event_seq FROM board_events;

DROP TABLE board_events;
ALTER TABLE board_events_flat RENAME TO board_events;

CREATE INDEX idx_board_events_board_created ON board_events(board_id, created_at DESC);
CREATE INDEX idx_board_events_board_seq ON board_events(board_id, event_seq DESC);
