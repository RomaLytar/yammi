-- Откат миграции 000001_init

-- Удаляем таблицу cards (партиции удалятся автоматически)
DROP TABLE IF EXISTS cards CASCADE;

-- Удаляем таблицу columns
DROP TABLE IF EXISTS columns CASCADE;

-- Удаляем таблицу board_members
DROP TABLE IF EXISTS board_members CASCADE;

-- Удаляем таблицу boards
DROP TABLE IF EXISTS boards CASCADE;
