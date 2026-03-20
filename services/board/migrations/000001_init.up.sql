-- Boards table (метаданные)
CREATE TABLE boards (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    owner_id UUID NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_boards_owner_id ON boards(owner_id);
CREATE INDEX idx_boards_cursor ON boards(created_at DESC, id DESC);

-- Board members (sharing, many-to-many)
CREATE TABLE board_members (
    board_id UUID REFERENCES boards(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('owner', 'member')),
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (board_id, user_id)
);

CREATE INDEX idx_board_members_user_id ON board_members(user_id);

-- Columns
CREATE TABLE columns (
    id UUID PRIMARY KEY,
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    position INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_columns_board_id ON columns(board_id);
CREATE INDEX idx_columns_position ON columns(board_id, position);

-- Cards (partitioned по board_id для performance)
CREATE TABLE cards (
    id UUID NOT NULL,
    column_id UUID NOT NULL,
    board_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    position VARCHAR(100) NOT NULL,
    assignee_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (board_id, id)
) PARTITION BY HASH (board_id);

-- Создаем 4 партиции
CREATE TABLE cards_p0 PARTITION OF cards FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE cards_p1 PARTITION OF cards FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE cards_p2 PARTITION OF cards FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE cards_p3 PARTITION OF cards FOR VALUES WITH (MODULUS 4, REMAINDER 3);

-- Индексы для cards (создаются для каждой партиции автоматически)
CREATE INDEX idx_cards_column_id ON cards(column_id);
CREATE INDEX idx_cards_position ON cards(column_id, position);
CREATE INDEX idx_cards_assignee_id ON cards(assignee_id) WHERE assignee_id IS NOT NULL;
