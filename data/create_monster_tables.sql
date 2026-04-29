CREATE TABLE IF NOT EXISTS save_monsters (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    room_db_id          INTEGER NOT NULL REFERENCES save_rooms(id) ON DELETE CASCADE,
    monster_template_id TEXT NOT NULL,
    is_defeated         BOOLEAN NOT NULL DEFAULT 0,
    level               INTEGER NOT NULL DEFAULT 1,
    current_xp          INTEGER NOT NULL DEFAULT 0,
    current_hp          INTEGER NOT NULL DEFAULT 0,
    current_mana        INTEGER NOT NULL DEFAULT 0,
    lb_health           INTEGER NOT NULL DEFAULT 0,
    lb_mana             INTEGER NOT NULL DEFAULT 0,
    lb_attack           INTEGER NOT NULL DEFAULT 0,
    lb_defense          INTEGER NOT NULL DEFAULT 0,
    lb_magic            INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS save_monster_status_effects (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    monster_db_id     INTEGER NOT NULL REFERENCES save_monsters(id) ON DELETE CASCADE,
    effect_type       TEXT NOT NULL,
    stat_affected     TEXT NOT NULL,
    delta             INTEGER NOT NULL DEFAULT 0,
    duration          INTEGER NOT NULL DEFAULT 0,
    target            TEXT NOT NULL,
    activation_delay  INTEGER NOT NULL DEFAULT 0,
    turns_remaining   INTEGER NOT NULL DEFAULT 0,
    turns_to_activate INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS save_room_events (
    room_db_id        INTEGER PRIMARY KEY REFERENCES save_rooms(id) ON DELETE CASCADE,
    event_template_id TEXT NOT NULL,
    applied           BOOLEAN NOT NULL DEFAULT 0
);
