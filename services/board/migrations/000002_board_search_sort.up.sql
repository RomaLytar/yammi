CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Trigram index for fast ILIKE search by title
CREATE INDEX idx_boards_title_trgm ON boards USING gin (title gin_trgm_ops);

-- Index for sorting by updated_at
CREATE INDEX idx_boards_updated_at ON boards(updated_at DESC, id DESC);
