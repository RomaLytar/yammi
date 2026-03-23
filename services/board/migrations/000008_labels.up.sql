CREATE TABLE IF NOT EXISTS labels (
    id UUID PRIMARY KEY,
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6b7280',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(board_id, name)
);
CREATE INDEX IF NOT EXISTS idx_labels_board_id ON labels(board_id);

CREATE TABLE IF NOT EXISTS card_labels (
    card_id UUID NOT NULL,
    board_id UUID NOT NULL,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (board_id, card_id, label_id)
) PARTITION BY HASH (board_id);

CREATE TABLE IF NOT EXISTS card_labels_p0 PARTITION OF card_labels FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE IF NOT EXISTS card_labels_p1 PARTITION OF card_labels FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE IF NOT EXISTS card_labels_p2 PARTITION OF card_labels FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE IF NOT EXISTS card_labels_p3 PARTITION OF card_labels FOR VALUES WITH (MODULUS 4, REMAINDER 3);

CREATE INDEX IF NOT EXISTS idx_card_labels_card_id ON card_labels(card_id, board_id);
CREATE INDEX IF NOT EXISTS idx_card_labels_label_id ON card_labels(label_id);
