-- Шаблоны карточек
CREATE TABLE IF NOT EXISTS card_templates (
    id UUID PRIMARY KEY,
    board_id UUID REFERENCES boards(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    task_type VARCHAR(20) NOT NULL DEFAULT 'task',
    checklist_data JSONB NOT NULL DEFAULT '[]',
    label_ids UUID[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_card_templates_board ON card_templates(board_id) WHERE board_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_card_templates_user ON card_templates(user_id);

-- Шаблоны колонок (списков)
CREATE TABLE IF NOT EXISTS column_templates (
    id UUID PRIMARY KEY,
    board_id UUID REFERENCES boards(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    columns_data JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_column_templates_board ON column_templates(board_id) WHERE board_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_column_templates_user ON column_templates(user_id);

-- Шаблоны досок
CREATE TABLE IF NOT EXISTS board_templates (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    columns_data JSONB NOT NULL DEFAULT '[]',
    labels_data JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_board_templates_user ON board_templates(user_id);
