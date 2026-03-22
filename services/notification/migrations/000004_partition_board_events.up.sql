-- Партиционирование board_events по board_id (HASH, 8 партиций)
-- Аналогично cards в board service

CREATE TABLE board_events_partitioned (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL,
    actor_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (board_id, id)
) PARTITION BY HASH (board_id);

CREATE TABLE board_events_p0 PARTITION OF board_events_partitioned FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE board_events_p1 PARTITION OF board_events_partitioned FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE board_events_p2 PARTITION OF board_events_partitioned FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE board_events_p3 PARTITION OF board_events_partitioned FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE board_events_p4 PARTITION OF board_events_partitioned FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE board_events_p5 PARTITION OF board_events_partitioned FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE board_events_p6 PARTITION OF board_events_partitioned FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE board_events_p7 PARTITION OF board_events_partitioned FOR VALUES WITH (MODULUS 8, REMAINDER 7);

-- Копируем существующие данные
INSERT INTO board_events_partitioned (id, board_id, actor_id, event_type, title, message, metadata, created_at)
SELECT id, board_id, actor_id, event_type, title, message, metadata, created_at FROM board_events;

-- Подменяем таблицы
DROP TABLE board_events;
ALTER TABLE board_events_partitioned RENAME TO board_events;

-- Индексы (создаются на каждой партиции автоматически)
CREATE INDEX idx_board_events_board_created ON board_events(board_id, created_at DESC);
