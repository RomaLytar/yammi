DROP INDEX IF EXISTS idx_cards_creator_id;
ALTER TABLE cards DROP COLUMN IF EXISTS creator_id;
