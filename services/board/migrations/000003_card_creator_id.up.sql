ALTER TABLE cards ADD COLUMN creator_id UUID;

-- Set existing cards' creator_id to board owner for backward compatibility
UPDATE cards c SET creator_id = b.owner_id FROM boards b WHERE c.board_id = b.id AND c.creator_id IS NULL;

ALTER TABLE cards ALTER COLUMN creator_id SET NOT NULL;

CREATE INDEX idx_cards_creator_id ON cards(creator_id);
