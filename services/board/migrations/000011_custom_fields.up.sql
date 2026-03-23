CREATE TABLE IF NOT EXISTS custom_field_definitions (
    id UUID PRIMARY KEY,
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    field_type VARCHAR(20) NOT NULL,
    options JSONB,
    position INT NOT NULL DEFAULT 0,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(board_id, name),
    CONSTRAINT chk_field_type CHECK (field_type IN ('text', 'number', 'date', 'dropdown'))
);
CREATE INDEX IF NOT EXISTS idx_custom_field_defs_board ON custom_field_definitions(board_id, position);

CREATE TABLE IF NOT EXISTS custom_field_values (
    id UUID NOT NULL,
    card_id UUID NOT NULL,
    board_id UUID NOT NULL,
    field_id UUID NOT NULL REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
    value_text TEXT,
    value_number DOUBLE PRECISION,
    value_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (board_id, id),
    UNIQUE (board_id, card_id, field_id)
) PARTITION BY HASH (board_id);

CREATE TABLE IF NOT EXISTS custom_field_values_p0 PARTITION OF custom_field_values FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE IF NOT EXISTS custom_field_values_p1 PARTITION OF custom_field_values FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE IF NOT EXISTS custom_field_values_p2 PARTITION OF custom_field_values FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE IF NOT EXISTS custom_field_values_p3 PARTITION OF custom_field_values FOR VALUES WITH (MODULUS 4, REMAINDER 3);

CREATE INDEX IF NOT EXISTS idx_custom_field_values_card ON custom_field_values(card_id, board_id);
CREATE INDEX IF NOT EXISTS idx_custom_field_values_field ON custom_field_values(field_id, board_id);
