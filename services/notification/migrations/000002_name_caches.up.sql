-- Кеш имён досок (из board.created / board.updated событий)
CREATE TABLE board_names (
    board_id UUID PRIMARY KEY,
    title VARCHAR(500) NOT NULL
);

-- Кеш имён пользователей (из user.created событий)
CREATE TABLE user_names (
    user_id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- Кеш имён карточек (из card.created / card.updated событий)
CREATE TABLE card_names (
    card_id UUID PRIMARY KEY,
    title VARCHAR(500) NOT NULL
);

-- Кеш имён колонок (из column.created / column.updated событий)
CREATE TABLE column_names (
    column_id UUID PRIMARY KEY,
    title VARCHAR(500) NOT NULL
);
