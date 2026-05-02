package database

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"tbrpg/game"
	"tbrpg/models"
)




// this file handles saving updating fetching and deleting games from the database





func (db *DataBase) ListSaves() ([]Save, error) {
	rows, err := db.Data.Query(`SELECT id, COALESCE(label, ''), saved_at FROM saves ORDER BY saved_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	saves := []Save{}
	for rows.Next() {
		var s Save
		if err := rows.Scan(&s.Id, &s.Label, &s.SavedAt); err != nil {
			return nil, err
		}
		saves = append(saves, s)
	}
	return saves, nil
}

func (db *DataBase) DeleteSave(id int) error {
	_, err := db.Data.Exec(`DELETE FROM saves WHERE id = ?`, id)
	return err
}

func (db *DataBase) CreateGame(g *game.Game) (int, error) {
	tx, err := db.Data.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO saves (is_endless, is_in_battle, current_room_id) VALUES (?, ?, ?)`,
		g.IsEndless, g.IsInBattle, g.CurrentRoomId,
	)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	id := int(id64)

	if err := writeSaveChildren(tx, id, g); err != nil {
		return 0, err
	}

	return id, tx.Commit()
}

func (db *DataBase) SaveGame(g *game.Game) error {
	tx, err := db.Data.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE saves SET is_endless = ?, is_in_battle = ?, current_room_id = ?, saved_at = CURRENT_TIMESTAMP WHERE id = ?`,
		g.IsEndless, g.IsInBattle, g.CurrentRoomId, g.Id,
	)
	if err != nil {
		return err
	}

	// Delete child rows. Rooms/monsters/events cascade from floors.
	childTables := []string{
		"save_hero_status_effects",
		"save_hero_item_pool",
		"save_hero_equipped_items",
		"save_hero_equipped_moves",
		"save_hero_learned_moves",
		"save_heroes",
		"save_pending_level_ups",
		"save_shop_items",
	}
	for _, t := range childTables {
		if _, err := tx.Exec(`DELETE FROM `+t+` WHERE save_id = ?`, g.Id); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM save_floors WHERE save_id = ?`, g.Id); err != nil {
		return err
	}

	if err := writeSaveChildren(tx, g.Id, g); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DataBase) LoadSave(id int, config *game.GameConfig) (*game.Game, error) {
	g := &game.Game{}

	var isEndless, isInBattle bool
	err := db.Data.QueryRow(
		`SELECT id, is_endless, is_in_battle, current_room_id FROM saves WHERE id = ?`, id,
	).Scan(&g.Id, &isEndless, &isInBattle, &g.CurrentRoomId)
	if err != nil {
		return nil, err
	}
	g.IsEndless = isEndless
	g.IsInBattle = isInBattle

	hero, err := readHero(db.Data, id, config)
	if err != nil {
		return nil, err
	}
	g.Player = *hero

	floors, err := readFloors(db.Data, id, config)
	if err != nil {
		return nil, err
	}
	g.Floors = floors

	shop, err := readShop(db.Data, id, config)
	if err != nil {
		return nil, err
	}
	g.Shop = shop

	var manual, random int
	err = db.Data.QueryRow(
		`SELECT manual_points, random_points FROM save_pending_level_ups WHERE save_id = ?`, id,
	).Scan(&manual, &random)
	if err == nil {
		g.PendingLevelUp = &game.PendingAllocation{ManualPoints: manual, RandomPoints: random}
	}

	g.AllMoves = config.Moves
	g.AllItems = config.Items
	g.Settings = config.Settings
	g.Config = config

	return g, nil
}

// writeSaveChildren inserts all child records for a save inside a transaction.
func writeSaveChildren(tx *sql.Tx, saveId int, g *game.Game) error {
	if err := writeHero(tx, saveId, &g.Player); err != nil {
		return err
	}
	if err := writeFloors(tx, saveId, g.Floors); err != nil {
		return err
	}
	if err := writeShop(tx, saveId, g.Shop); err != nil {
		return err
	}
	if err := writePendingLevelUp(tx, saveId, g.PendingLevelUp); err != nil {
		return err
	}
	return nil
}

// --- write helpers ---

func writeHero(tx *sql.Tx, saveId int, h *models.Hero) error {
	if _, err := tx.Exec(
		`INSERT INTO save_heroes (save_id, hero_template_id, level, current_xp, current_hp, current_mana, current_gold, lb_health, lb_mana, lb_attack, lb_defense, lb_magic)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		saveId, h.Id, h.Level, h.CurrentXP, h.CurrentHP, h.CurrentMana, h.CurrentGold,
		h.LevelStats.Health, h.LevelStats.Mana, h.LevelStats.Attack, h.LevelStats.Defense, h.LevelStats.Magic,
	); err != nil {
		return err
	}

	for _, lm := range h.LearnedMoves {
		if _, err := tx.Exec(
			`INSERT INTO save_hero_learned_moves (save_id, move_id, level) VALUES (?, ?, ?)`,
			saveId, lm.MoveId, lm.Level,
		); err != nil {
			return err
		}
	}

	for slot, moveId := range h.EquippedMoves {
		if _, err := tx.Exec(
			`INSERT INTO save_hero_equipped_moves (save_id, move_id, slot_order) VALUES (?, ?, ?)`,
			saveId, moveId, slot,
		); err != nil {
			return err
		}
	}

	for _, item := range h.EquippedItems {
		if _, err := tx.Exec(
			`INSERT INTO save_hero_equipped_items (save_id, item_id) VALUES (?, ?)`,
			saveId, item.Id,
		); err != nil {
			return err
		}
	}

	for _, itemId := range h.ItemPool {
		if _, err := tx.Exec(
			`INSERT INTO save_hero_item_pool (save_id, item_id) VALUES (?, ?)`,
			saveId, itemId,
		); err != nil {
			return err
		}
	}

	for _, se := range h.StatusEffects {
		if _, err := tx.Exec(
			`INSERT INTO save_hero_status_effects (save_id, effect_type, stat_affected, delta, duration, target, activation_delay, turns_remaining, turns_to_activate)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			saveId, string(se.Type), string(se.StatAffected), se.BaseDelta, se.Duration, string(se.Target),
			se.ActivationDelay, se.TurnsRemaining, se.TurnsToActivate,
		); err != nil {
			return err
		}
	}

	return nil
}

func writeFloors(tx *sql.Tx, saveId int, floors []models.Floor) error {
	for _, floor := range floors {
		res, err := tx.Exec(
			`INSERT INTO save_floors (save_id, floor_idx, is_completed) VALUES (?, ?, ?)`,
			saveId, floor.Idx, floor.IsCompleted,
		)
		if err != nil {
			return err
		}
		floorDBId, _ := res.LastInsertId()

		for _, room := range floor.Rooms {
			nextIds := strings.Join(room.NextRoomIds, ";")
			res2, err := tx.Exec(
				`INSERT INTO save_rooms (floor_db_id, room_string_id, encounter_kind, is_completed, can_enter, next_room_ids, environment_id)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`,
				floorDBId, room.Id, string(room.Encounter.Kind), room.IsCompleted, room.CanEnter, nextIds, room.Environment.Id,
			)
			if err != nil {
				return err
			}
			roomDBId, _ := res2.LastInsertId()

			switch room.Encounter.Kind {
			case models.EncounterKindMonster, models.EncounterKindBoss:
				if room.Encounter.Monster != nil {
					if err := writeMonster(tx, roomDBId, room.Encounter.Monster); err != nil {
						return err
					}
				}
			case models.EncounterKindEvent:
				if room.Encounter.Event != nil {
					if _, err := tx.Exec(
						`INSERT INTO save_room_events (room_db_id, event_template_id, applied) VALUES (?, ?, ?)`,
						roomDBId, room.Encounter.Event.Id, room.Encounter.Event.Applied,
					); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func writeMonster(tx *sql.Tx, roomDBId int64, m *models.Monster) error {
	res, err := tx.Exec(
		`INSERT INTO save_monsters (room_db_id, monster_template_id, is_defeated, level, current_xp, current_hp, current_mana, lb_health, lb_mana, lb_attack, lb_defense, lb_magic)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		roomDBId, m.Id, false, m.Level, m.CurrentXP, m.CurrentHP, m.CurrentMana,
		m.LevelStats.Health, m.LevelStats.Mana, m.LevelStats.Attack, m.LevelStats.Defense, m.LevelStats.Magic,
	)
	if err != nil {
		return err
	}
	monsterDBId, _ := res.LastInsertId()

	for _, se := range m.StatusEffects {
		if _, err := tx.Exec(
			`INSERT INTO save_monster_status_effects (monster_db_id, effect_type, stat_affected, delta, duration, target, activation_delay, turns_remaining, turns_to_activate)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			monsterDBId, string(se.Type), string(se.StatAffected), se.BaseDelta, se.Duration, string(se.Target),
			se.ActivationDelay, se.TurnsRemaining, se.TurnsToActivate,
		); err != nil {
			return err
		}
	}
	return nil
}

func writeShop(tx *sql.Tx, saveId int, shop *models.Shop) error {
	if shop == nil {
		return nil
	}
	for _, item := range shop.Items {
		if _, err := tx.Exec(
			`INSERT INTO save_shop_items (save_id, item_id) VALUES (?, ?)`,
			saveId, item.Id,
		); err != nil {
			return err
		}
	}
	return nil
}

func writePendingLevelUp(tx *sql.Tx, saveId int, p *game.PendingAllocation) error {
	if p == nil {
		return nil
	}
	_, err := tx.Exec(
		`INSERT INTO save_pending_level_ups (save_id, manual_points, random_points) VALUES (?, ?, ?)`,
		saveId, p.ManualPoints, p.RandomPoints,
	)
	return err
}

// --- read helpers ---

func readHero(db *sql.DB, saveId int, config *game.GameConfig) (*models.Hero, error) {
	var templateId string
	var level, xp, hp, mana, gold int
	var lbH, lbMa, lbAt, lbDe, lbMg int

	err := db.QueryRow(
		`SELECT hero_template_id, level, current_xp, current_hp, current_mana, current_gold,
		        lb_health, lb_mana, lb_attack, lb_defense, lb_magic
		 FROM save_heroes WHERE save_id = ?`, saveId,
	).Scan(&templateId, &level, &xp, &hp, &mana, &gold, &lbH, &lbMa, &lbAt, &lbDe, &lbMg)
	if err != nil {
		return nil, err
	}

	// Start from the config template to get base stats, scaling, env effects, etc.
	var hero models.Hero
	for _, t := range config.HeroTemplates {
		if t.Id == templateId {
			hero = t
			break
		}
	}
	if hero.Id == "" {
		return nil, fmt.Errorf("hero template %q not found in config", templateId)
	}

	hero.Level = level
	hero.CurrentXP = xp
	hero.CurrentHP = hp
	hero.CurrentMana = mana
	hero.CurrentGold = gold
	hero.LevelStats = models.Stats{
		Health:  lbH,
		Mana:    lbMa,
		Attack:  lbAt,
		Defense: lbDe,
		Magic:   lbMg,
	}

	// Learned moves
	rows, err := db.Query(
		`SELECT move_id, level FROM save_hero_learned_moves WHERE save_id = ?`, saveId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hero.LearnedMoves = nil
	for rows.Next() {
		var lm models.LearnedMove
		if err := rows.Scan(&lm.MoveId, &lm.Level); err != nil {
			return nil, err
		}
		hero.LearnedMoves = append(hero.LearnedMoves, lm)
	}

	// Equipped moves (ordered by slot)
	rows2, err := db.Query(
		`SELECT move_id FROM save_hero_equipped_moves WHERE save_id = ? ORDER BY slot_order`, saveId,
	)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	hero.EquippedMoves = nil
	for rows2.Next() {
		var moveId string
		if err := rows2.Scan(&moveId); err != nil {
			return nil, err
		}
		hero.EquippedMoves = append(hero.EquippedMoves, moveId)
	}

	// Equipped items
	rows3, err := db.Query(
		`SELECT item_id FROM save_hero_equipped_items WHERE save_id = ?`, saveId,
	)
	if err != nil {
		return nil, err
	}
	defer rows3.Close()
	hero.EquippedItems = nil
	for rows3.Next() {
		var itemId string
		if err := rows3.Scan(&itemId); err != nil {
			return nil, err
		}
		if item, ok := config.Items[itemId]; ok {
			hero.EquippedItems = append(hero.EquippedItems, item)
		}
	}

	// Item pool
	rows4, err := db.Query(
		`SELECT item_id FROM save_hero_item_pool WHERE save_id = ?`, saveId,
	)
	if err != nil {
		return nil, err
	}
	defer rows4.Close()
	hero.ItemPool = nil
	for rows4.Next() {
		var itemId string
		if err := rows4.Scan(&itemId); err != nil {
			return nil, err
		}
		hero.ItemPool = append(hero.ItemPool, itemId)
	}

	// Status effects
	rows5, err := db.Query(
		`SELECT effect_type, stat_affected, delta, duration, target, activation_delay, turns_remaining, turns_to_activate
		 FROM save_hero_status_effects WHERE save_id = ?`, saveId,
	)
	if err != nil {
		return nil, err
	}
	defer rows5.Close()
	hero.StatusEffects = nil
	for rows5.Next() {
		var se models.StatusEffect
		var effectType, statAffected, target string
		if err := rows5.Scan(&effectType, &statAffected, &se.BaseDelta, &se.Duration, &target, &se.ActivationDelay, &se.TurnsRemaining, &se.TurnsToActivate); err != nil {
			return nil, err
		}
		se.Type = models.EffectType(effectType)
		se.StatAffected = models.StatType(statAffected)
		se.Target = models.EffectTarget(target)
		hero.StatusEffects = append(hero.StatusEffects, se)
	}

	return &hero, nil
}

func readFloors(db *sql.DB, saveId int, config *game.GameConfig) ([]models.Floor, error) {
	floorRows, err := db.Query(
		`SELECT id, floor_idx, is_completed FROM save_floors WHERE save_id = ? ORDER BY floor_idx`, saveId,
	)
	if err != nil {
		return nil, err
	}
	defer floorRows.Close()

	var floors []models.Floor
	for floorRows.Next() {
		var floorDBId int
		var floor models.Floor
		if err := floorRows.Scan(&floorDBId, &floor.Idx, &floor.IsCompleted); err != nil {
			return nil, err
		}

		rooms, err := readRooms(db, floorDBId, config)
		if err != nil {
			return nil, err
		}
		floor.Rooms = rooms
		floors = append(floors, floor)
	}
	return floors, nil
}

func readRooms(db *sql.DB, floorDBId int, config *game.GameConfig) ([]models.Room, error) {
	roomRows, err := db.Query(
		`SELECT id, room_string_id, encounter_kind, is_completed, can_enter, next_room_ids, environment_id
		 FROM save_rooms WHERE floor_db_id = ?`, floorDBId,
	)
	if err != nil {
		return nil, err
	}
	defer roomRows.Close()

	var rooms []models.Room
	for roomRows.Next() {
		var roomDBId int
		var room models.Room
		var encounterKind, nextRoomIdsRaw, environmentId string
		if err := roomRows.Scan(&roomDBId, &room.Id, &encounterKind, &room.IsCompleted, &room.CanEnter, &nextRoomIdsRaw, &environmentId); err != nil {
			return nil, err
		}

		room.Encounter.Kind = models.EncounterKind(encounterKind)

		if nextRoomIdsRaw != "" {
			room.NextRoomIds = strings.Split(nextRoomIdsRaw, ";")
		}

		// Resolve environment from config
		for _, env := range config.EnvironmentTemplates {
			if env.Id == environmentId {
				room.Environment = env
				break
			}
		}

		switch room.Encounter.Kind {
		case models.EncounterKindMonster, models.EncounterKindBoss:
			monster, err := readMonster(db, roomDBId, config)
			if err != nil {
				return nil, err
			}
			room.Encounter.Monster = monster
		case models.EncounterKindEvent:
			event, err := readEvent(db, roomDBId, config)
			if err != nil {
				return nil, err
			}
			room.Encounter.Event = event
		}

		rooms = append(rooms, room)
	}
	return rooms, nil
}

func readMonster(db *sql.DB, roomDBId int, config *game.GameConfig) (*models.Monster, error) {
	var monsterDBId int
	var templateId string
	var level, xp, hp, mana int
	var lbH, lbMa, lbAt, lbDe, lbMg int

	err := db.QueryRow(
		`SELECT id, monster_template_id, is_defeated, level, current_xp, current_hp, current_mana,
		        lb_health, lb_mana, lb_attack, lb_defense, lb_magic
		 FROM save_monsters WHERE room_db_id = ?`, roomDBId,
	).Scan(&monsterDBId, &templateId, new(bool), &level, &xp, &hp, &mana,
		&lbH, &lbMa, &lbAt, &lbDe, &lbMg)
	if err != nil {
		return nil, err
	}

	// Start from the matching template (checks both monsters and bosses)
	var monster models.Monster
	for _, t := range config.MonsterTemplates {
		if t.Id == templateId {
			monster = t
			break
		}
	}
	if monster.Id == "" {
		for _, t := range config.BossTemplates {
			if t.Id == templateId {
				monster = t
				break
			}
		}
	}
	if monster.Id == "" {
		return nil, fmt.Errorf("monster template %q not found in config", templateId)
	}

	monster.Level = level
	monster.CurrentXP = xp
	monster.CurrentHP = hp
	monster.CurrentMana = mana
	monster.LevelStats = models.Stats{
		Health:  lbH,
		Mana:    lbMa,
		Attack:  lbAt,
		Defense: lbDe,
		Magic:   lbMg,
	}

	// Status effects
	rows, err := db.Query(
		`SELECT effect_type, stat_affected, delta, duration, target, activation_delay, turns_remaining, turns_to_activate
		 FROM save_monster_status_effects WHERE monster_db_id = ?`, monsterDBId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var se models.StatusEffect
		var effectType, statAffected, target string
		if err := rows.Scan(&effectType, &statAffected, &se.BaseDelta, &se.Duration, &target, &se.ActivationDelay, &se.TurnsRemaining, &se.TurnsToActivate); err != nil {
			return nil, err
		}
		se.Type = models.EffectType(effectType)
		se.StatAffected = models.StatType(statAffected)
		se.Target = models.EffectTarget(target)
		monster.StatusEffects = append(monster.StatusEffects, se)
	}

	return &monster, nil
}

func readEvent(db *sql.DB, roomDBId int, config *game.GameConfig) (*models.Event, error) {
	var eventTemplateId string
	var applied bool
	err := db.QueryRow(
		`SELECT event_template_id, applied FROM save_room_events WHERE room_db_id = ?`, roomDBId,
	).Scan(&eventTemplateId, &applied)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(config.EventTemplates, func(e models.Event) bool { return e.Id == eventTemplateId })
	if idx == -1 {
		return nil, nil
	}
	event := config.EventTemplates[idx]
	event.Applied = applied
	return &event, nil
}

func readShop(db *sql.DB, saveId int, config *game.GameConfig) (*models.Shop, error) {
	rows, err := db.Query(
		`SELECT item_id FROM save_shop_items WHERE save_id = ?`, saveId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var itemId string
		if err := rows.Scan(&itemId); err != nil {
			return nil, err
		}
		if item, ok := config.Items[itemId]; ok {
			items = append(items, item)
		}
	}

	if items == nil {
		return nil, nil
	}
	return &models.Shop{Items: items}, nil
}
