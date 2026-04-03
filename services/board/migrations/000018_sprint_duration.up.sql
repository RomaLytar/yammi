-- Sprint duration (days) for auto-calculating release end date. Default 14 (2 weeks). Min 7.
ALTER TABLE board_settings ADD COLUMN IF NOT EXISTS sprint_duration_days INT NOT NULL DEFAULT 14;
