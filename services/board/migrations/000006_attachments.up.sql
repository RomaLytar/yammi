CREATE TABLE attachments (
    id UUID NOT NULL,
    card_id UUID NOT NULL,
    board_id UUID NOT NULL,
    file_name VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    storage_key VARCHAR(1000) NOT NULL,
    uploader_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (board_id, id)
) PARTITION BY HASH (board_id);

CREATE TABLE attachments_p0 PARTITION OF attachments FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE attachments_p1 PARTITION OF attachments FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE attachments_p2 PARTITION OF attachments FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE attachments_p3 PARTITION OF attachments FOR VALUES WITH (MODULUS 4, REMAINDER 3);

CREATE INDEX idx_attachments_card_id ON attachments(card_id, board_id, created_at DESC);
CREATE INDEX idx_attachments_uploader_id ON attachments(uploader_id) WHERE uploader_id IS NOT NULL;
