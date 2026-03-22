CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL,
    board_id UUID NOT NULL,
    author_id UUID NOT NULL,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL CHECK (length(content) > 0 AND length(content) <= 10000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_card_id ON comments(card_id, created_at ASC);
CREATE INDEX idx_comments_parent_id ON comments(parent_id, created_at ASC) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_comments_author_id ON comments(author_id, created_at DESC);
