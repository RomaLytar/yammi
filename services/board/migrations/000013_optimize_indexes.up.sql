-- Улучшенный индекс для card_links: child_id queries без board_id (ListParents)
-- Существующий idx_card_links_child(child_id) расширяем board_id для partition pruning
DROP INDEX IF EXISTS idx_card_links_child;
CREATE INDEX idx_card_links_child ON card_links(child_id, board_id);

-- Индекс для custom_field_values с ORDER BY created_at
DROP INDEX IF EXISTS idx_custom_field_values_card;
CREATE INDEX idx_custom_field_values_card ON custom_field_values(card_id, board_id, created_at);

-- Индекс для card_labels cleanup по board_id (при удалении доски)
CREATE INDEX IF NOT EXISTS idx_card_labels_board ON card_labels(board_id);
