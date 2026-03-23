ALTER TABLE cards ADD COLUMN due_date TIMESTAMPTZ;
ALTER TABLE cards ADD COLUMN priority VARCHAR(20) NOT NULL DEFAULT 'medium';
ALTER TABLE cards ADD COLUMN task_type VARCHAR(20) NOT NULL DEFAULT 'task';
CREATE INDEX idx_cards_due_date ON cards(due_date) WHERE due_date IS NOT NULL;
CREATE INDEX idx_cards_priority ON cards(board_id, priority);
ALTER TABLE cards ADD CONSTRAINT chk_priority CHECK (priority IN ('low', 'medium', 'high', 'critical'));
ALTER TABLE cards ADD CONSTRAINT chk_task_type CHECK (task_type IN ('bug', 'feature', 'task', 'improvement'));
