CREATE TABLE IF NOT EXISTS save_floors (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    save_id      INTEGER NOT NULL REFERENCES saves(id) ON DELETE CASCADE,
    floor_idx    INTEGER NOT NULL,
    is_completed BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS save_rooms (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    floor_db_id    INTEGER NOT NULL REFERENCES save_floors(id) ON DELETE CASCADE,
    room_string_id TEXT NOT NULL,
    encounter_kind TEXT NOT NULL,
    is_completed   BOOLEAN NOT NULL DEFAULT 0,
    can_enter      BOOLEAN NOT NULL DEFAULT 0,
    next_room_ids  TEXT NOT NULL DEFAULT '',
    environment_id TEXT NOT NULL DEFAULT ''
);
