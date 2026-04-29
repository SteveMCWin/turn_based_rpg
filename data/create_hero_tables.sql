CREATE TABLE IF NOT EXISTS save_heroes (
    save_id          INTEGER PRIMARY KEY REFERENCES saves(id) ON DELETE CASCADE,
    hero_template_id TEXT NOT NULL,
    level            INTEGER NOT NULL DEFAULT 1,
    current_xp       INTEGER NOT NULL DEFAULT 0,
    current_hp       INTEGER NOT NULL DEFAULT 0,
    current_mana     INTEGER NOT NULL DEFAULT 0,
    current_gold     INTEGER NOT NULL DEFAULT 0,
    lb_health        INTEGER NOT NULL DEFAULT 0,
    lb_mana          INTEGER NOT NULL DEFAULT 0,
    lb_attack        INTEGER NOT NULL DEFAULT 0,
    lb_defense       INTEGER NOT NULL DEFAULT 0,
    lb_magic         INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS save_hero_learned_moves (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    save_id   INTEGER NOT NULL REFERENCES saves(id) ON DELETE CASCADE,
    move_id   TEXT NOT NULL,
    level     INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS save_hero_equipped_moves (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    save_id    INTEGER NOT NULL REFERENCES saves(id) ON DELETE CASCADE,
    move_id    TEXT NOT NULL,
    slot_order INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS save_hero_equipped_items (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    save_id INTEGER NOT NULL REFERENCES saves(id) ON DELETE CASCADE,
    item_id TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS save_hero_item_pool (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    save_id INTEGER NOT NULL REFERENCES saves(id) ON DELETE CASCADE,
    item_id TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS save_hero_status_effects (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    save_id           INTEGER NOT NULL REFERENCES saves(id) ON DELETE CASCADE,
    effect_type       TEXT NOT NULL,
    stat_affected     TEXT NOT NULL,
    delta             INTEGER NOT NULL DEFAULT 0,
    duration          INTEGER NOT NULL DEFAULT 0,
    target            TEXT NOT NULL,
    activation_delay  INTEGER NOT NULL DEFAULT 0,
    turns_remaining   INTEGER NOT NULL DEFAULT 0,
    turns_to_activate INTEGER NOT NULL DEFAULT 0
);
