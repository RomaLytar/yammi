-- Board events: 1 строка на событие (вместо N строк в notifications при fan-out)
CREATE TABLE board_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL,
    actor_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_board_events_board_created ON board_events(board_id, created_at DESC);

-- Курсор "прочитано до" для каждого пользователя на каждой доске
CREATE TABLE user_board_cursors (
    user_id UUID NOT NULL,
    board_id UUID NOT NULL,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, board_id)
);
