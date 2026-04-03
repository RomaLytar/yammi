-- Планируемые даты начала и окончания релиза (пользователь выбирает при создании)
ALTER TABLE releases ADD COLUMN IF NOT EXISTS start_date TIMESTAMPTZ;
ALTER TABLE releases ADD COLUMN IF NOT EXISTS end_date TIMESTAMPTZ;
