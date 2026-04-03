-- Releases: board-scoped release management (draft → active → completed)
CREATE TABLE IF NOT EXISTS releases (
    id UUID PRIMARY KEY,
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'completed')),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_by UUID NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Only ONE active release per board (enforced at DB level)
CREATE UNIQUE INDEX idx_releases_board_active ON releases(board_id) WHERE status = 'active';
CREATE INDEX idx_releases_board_id ON releases(board_id);
CREATE INDEX idx_releases_board_status ON releases(board_id, status);

-- Cards can optionally belong to a release (NULL = backlog)
-- ALTER TABLE on partitioned table propagates to all partitions automatically
ALTER TABLE cards ADD COLUMN IF NOT EXISTS release_id UUID;
CREATE INDEX idx_cards_release_id ON cards(board_id, release_id) WHERE release_id IS NOT NULL;
CREATE INDEX idx_cards_backlog ON cards(board_id) WHERE release_id IS NULL;

-- Board settings: done column for release completion checks
ALTER TABLE board_settings ADD COLUMN IF NOT EXISTS done_column_id UUID;
