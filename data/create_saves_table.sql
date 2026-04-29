CREATE TABLE IF NOT EXISTS saves (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    label           TEXT,
    saved_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_endless      BOOLEAN NOT NULL DEFAULT 0,
    is_in_battle    BOOLEAN NOT NULL DEFAULT 0,
    current_room_id TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS save_pending_level_ups (
    save_id       INTEGER PRIMARY KEY REFERENCES saves(id) ON DELETE CASCADE,
    manual_points INTEGER NOT NULL DEFAULT 0,
    random_points INTEGER NOT NULL DEFAULT 0
);
