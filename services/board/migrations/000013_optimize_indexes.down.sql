DROP INDEX IF EXISTS idx_card_labels_board;

-- Восстанавливаем оригинальные индексы
DROP INDEX IF EXISTS idx_custom_field_values_card;
CREATE INDEX idx_custom_field_values_card ON custom_field_values(card_id, board_id);

DROP INDEX IF EXISTS idx_card_links_child;
CREATE INDEX idx_card_links_child ON card_links(child_id);
