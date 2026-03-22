CREATE TABLE card_activities (
    id UUID NOT NULL,
    card_id UUID NOT NULL,
    board_id UUID NOT NULL,
    actor_id UUID NOT NULL,
    activity_type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    changes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (board_id, id)
) PARTITION BY HASH (board_id);

CREATE TABLE card_activities_p0 PARTITION OF card_activities FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE card_activities_p1 PARTITION OF card_activities FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE card_activities_p2 PARTITION OF card_activities FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE card_activities_p3 PARTITION OF card_activities FOR VALUES WITH (MODULUS 4, REMAINDER 3);

CREATE INDEX idx_card_activities_card ON card_activities(card_id, board_id, created_at DESC);
