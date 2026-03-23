CREATE TABLE IF NOT EXISTS card_links (
    id UUID NOT NULL,
    parent_id UUID NOT NULL,
    child_id UUID NOT NULL,
    board_id UUID NOT NULL,
    link_type VARCHAR(20) NOT NULL DEFAULT 'subtask',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (board_id, id),
    UNIQUE (board_id, parent_id, child_id)
) PARTITION BY HASH (board_id);

CREATE TABLE IF NOT EXISTS card_links_p0 PARTITION OF card_links FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE IF NOT EXISTS card_links_p1 PARTITION OF card_links FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE IF NOT EXISTS card_links_p2 PARTITION OF card_links FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE IF NOT EXISTS card_links_p3 PARTITION OF card_links FOR VALUES WITH (MODULUS 4, REMAINDER 3);

CREATE INDEX IF NOT EXISTS idx_card_links_parent ON card_links(parent_id, board_id);
CREATE INDEX IF NOT EXISTS idx_card_links_child ON card_links(child_id);

ALTER TABLE card_links ADD CONSTRAINT chk_no_self_link CHECK (parent_id != child_id);
