CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}',
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_created ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_user_unread ON notifications(user_id) WHERE is_read = FALSE;
CREATE INDEX idx_notifications_user_type ON notifications(user_id, type, created_at DESC);
CREATE INDEX idx_notifications_search ON notifications USING gin (title gin_trgm_ops);

CREATE TABLE notification_settings (
    user_id UUID PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    realtime_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Local membership cache for routing notifications
CREATE TABLE board_members (
    board_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (board_id, user_id)
);
CREATE INDEX idx_board_members_user ON board_members(user_id);
