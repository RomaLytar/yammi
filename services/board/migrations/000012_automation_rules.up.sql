CREATE TABLE IF NOT EXISTS automation_rules (
    id UUID PRIMARY KEY,
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    trigger_type VARCHAR(50) NOT NULL,
    trigger_config JSONB NOT NULL DEFAULT '{}',
    action_type VARCHAR(50) NOT NULL,
    action_config JSONB NOT NULL DEFAULT '{}',
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_automation_rules_board ON automation_rules(board_id, enabled);
CREATE INDEX IF NOT EXISTS idx_automation_rules_trigger ON automation_rules(board_id, trigger_type) WHERE enabled = TRUE;

CREATE TABLE IF NOT EXISTS automation_executions (
    id UUID NOT NULL,
    rule_id UUID NOT NULL REFERENCES automation_rules(id) ON DELETE CASCADE,
    board_id UUID NOT NULL,
    card_id UUID,
    trigger_event_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    error_message TEXT,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (board_id, id)
) PARTITION BY HASH (board_id);

CREATE TABLE IF NOT EXISTS automation_executions_p0 PARTITION OF automation_executions FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE IF NOT EXISTS automation_executions_p1 PARTITION OF automation_executions FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE IF NOT EXISTS automation_executions_p2 PARTITION OF automation_executions FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE IF NOT EXISTS automation_executions_p3 PARTITION OF automation_executions FOR VALUES WITH (MODULUS 4, REMAINDER 3);

CREATE INDEX IF NOT EXISTS idx_automation_executions_rule ON automation_executions(rule_id, board_id, executed_at DESC);
