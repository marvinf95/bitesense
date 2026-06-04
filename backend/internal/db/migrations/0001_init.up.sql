PRAGMA foreign_keys = ON;

CREATE TABLE users (
    id          TEXT PRIMARY KEY,
    email       TEXT UNIQUE NOT NULL COLLATE NOCASE,
    pwd_hash    TEXT NOT NULL,
    locale      TEXT NOT NULL DEFAULT 'en',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE refresh_tokens (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  DATETIME NOT NULL,
    revoked_at  DATETIME,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX ix_refresh_user ON refresh_tokens(user_id);

CREATE TABLE meals (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    eaten_at     DATETIME NOT NULL,
    title        TEXT,
    notes        TEXT,
    source       TEXT NOT NULL CHECK(source IN ('text','image','barcode','favorite')),
    photo_path   TEXT,
    vision_raw   TEXT,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX ix_meals_user_eaten ON meals(user_id, eaten_at);

CREATE TABLE meal_items (
    id            TEXT PRIMARY KEY,
    meal_id       TEXT NOT NULL REFERENCES meals(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    display_name  TEXT NOT NULL,
    quantity_g    REAL,
    off_id        TEXT,
    confidence    REAL
);
CREATE INDEX ix_meal_items_name ON meal_items(name);
CREATE INDEX ix_meal_items_meal ON meal_items(meal_id);

CREATE TABLE meal_item_tags (
    meal_item_id  TEXT NOT NULL REFERENCES meal_items(id) ON DELETE CASCADE,
    tag           TEXT NOT NULL,
    PRIMARY KEY (meal_item_id, tag)
);
CREATE INDEX ix_meal_item_tags_tag ON meal_item_tags(tag);

CREATE TABLE symptoms (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    occurred_at    DATETIME NOT NULL,
    type           TEXT NOT NULL,
    severity       INTEGER NOT NULL CHECK(severity BETWEEN 1 AND 10),
    duration_min   INTEGER,
    bristol_stool  INTEGER CHECK(bristol_stool BETWEEN 1 AND 7),
    notes          TEXT,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX ix_symptoms_user_occurred ON symptoms(user_id, occurred_at);

CREATE TABLE meal_favorites (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label      TEXT NOT NULL,
    template   TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX ix_favorites_user ON meal_favorites(user_id);

CREATE TABLE vision_cache (
    image_hash    TEXT PRIMARY KEY,
    response_json TEXT NOT NULL,
    provider      TEXT NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sync_log (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    TEXT NOT NULL,
    entity     TEXT NOT NULL,
    entity_id  TEXT NOT NULL,
    op         TEXT NOT NULL CHECK(op IN ('insert','update','delete')),
    at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    payload    TEXT NOT NULL
);
CREATE INDEX ix_sync_user_at ON sync_log(user_id, at);
