package game

import (
	"fmt"
	"log"
	mrand "math/rand"
	"slices"

	"tbrpg/models"
)

// represents the whole game and it's state
type Game struct {
	ID                int                `json:"id"`
	Settings          GameSettings       `json:"settings"`
	Player            models.Hero        `json:"player"`
	Floors            []models.Floor     `json:"floors"`
	IsInBattle        bool               `json:"in_battle"`
	WaitingForMonster bool               `json:"waiting_for_monster"`
	IsEndless         bool               `json:"is_endless"`
	CurrentRoomID     string             `json:"current_room_id,omitempty"`
	LastBattleResult  *BattleResult      `json:"last_battle_result"`
	PendingLevelUp    *PendingAllocation `json:"pending_level_up"`

	// cached moves and items
	AllMoves map[string]models.MoveDefinition `json:"moves,omitempty"`
	AllItems map[string]models.Item           `json:"items,omitempty"`

	Shop *models.Shop `json:"shop,omitempty"`

	// dont' serialize to json
	Config *GameConfig `json:"-"`
}

// represents the amount of points that can be allocated to stats upon level up either manually or randomly
// note that the user can allocate more points randomly according ot the current config
type PendingAllocation struct {
	ManualPoints int `json:"manual_points"`
	RandomPoints int `json:"random_points"`
}

func NewGame(config *GameConfig, hero models.Hero) *Game {
	events := slices.Clone(config.EventTemplates)

	monsters := slices.Clone(config.MonsterTemplates)
	for i := range monsters {
		monsters[i].Reset()
	}

	bosses := slices.Clone(config.BossTemplates)
	for i := range bosses {
		bosses[i].Reset()
	}

	environments := slices.Clone(config.EnvironmentTemplates)

	g := Game{
		Settings: config.Settings,
		Player:   hero,
		Floors:   models.MakeFirstRealm(config.Settings.FloorsPerRealms, config.Settings.MaxRoomsPerLevel, config.Settings.MonsterSpawnChance),
		AllMoves: config.Moves,
		AllItems: config.Items,
		Shop:     models.NewShop(config.Items),
		Config:   config,
	}

	g.Player.Reset()
	models.FillFloorEncountersAndConnect(g.Floors, monsters, bosses, events, environments)

	return &g
}

// Utility
// pretty much just calls the floors AddRealmToExistingOne
func (g *Game) AddRealm(config *GameConfig) {
	monsters := slices.Clone(config.MonsterTemplates)
	for i := range monsters {
		monsters[i].Reset()
	}

	bosses := slices.Clone(config.BossTemplates)
	for i := range bosses {
		bosses[i].Reset()
	}

	g.Floors = models.AddRealmToExistingOne(
		g.Floors,
		config.Settings.FloorsPerRealms,
		config.Settings.MaxRoomsPerLevel,
		config.Settings.MonsterSpawnChance,
		monsters,
		bosses,
		slices.Clone(config.EventTemplates),
		config.EnvironmentTemplates,
	)
}

// Utility
func (g *Game) CurrentRoom() *models.Room {
	if g.CurrentRoomID == "" {
		return nil
	}
	return models.RoomByID(g.Floors, g.CurrentRoomID)
}

// The game handles applying environments of rooms to entities
func applyEnvironmentEffects(env models.Environment, monster *models.Monster, hero *models.Hero) {
	for env_id, effect := range monster.EnvironmentEffects {
		if env_id == env.Id {
			effect := effect
			addEnvironmentEffect(&effect, &monster.Entity)
		}
	}
	for env_id, effect := range hero.EnvironmentEffects {
		if env_id == env.Id {
			effect := effect
			addEnvironmentEffect(&effect, &hero.Entity)
		}
	}
}

// EnterRoom transitions the game into a room and returns any intro log lines.
// Note that upon entering a room of an already defeated enemy,
// The enemy may permanently level up so they are tougher to beat
func (g *Game) EnterRoom(roomID string) ([]string, error) {
	room := models.RoomByID(g.Floors, roomID)
	if room == nil {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}
	if !room.CanEnter {
		return nil, fmt.Errorf("room %s is not accessible", roomID)
	}

	var initialLog []string

	switch room.Encounter.Kind {
	case models.EncounterKindBoss:
		if room.IsCompleted {
			return nil, fmt.Errorf("boss already defeated")
		}

		room.Encounter.Monster.Reset()
		g.CurrentRoomID = roomID
		g.IsInBattle = true
		applyEnvironmentEffects(room.Environment, room.Encounter.Monster, &g.Player)

	case models.EncounterKindMonster:
		if room.IsCompleted {
			initialLog = append(initialLog, "You applied black magic to revive an already defeated foe. They are as hostile as you remember them to be.")
			if mrand.Int()%100 <= g.Settings.PercentChanceMonsterLevelsUp {
				room.Encounter.Monster.SetToLevel(room.Encounter.Monster.Level + 1)
				initialLog = append(initialLog, "However, upon reviving the monster for another duel, something went wrong in the ritual, and the monster is now permanently stronger!")
			}
		}

		room.Encounter.Monster.Reset()
		g.CurrentRoomID = roomID
		g.IsInBattle = true
		applyEnvironmentEffects(room.Environment, room.Encounter.Monster, &g.Player)

	case models.EncounterKindEvent:
		if room.Encounter.Event != nil && !room.Encounter.Event.Applied {
			g.applyEvent(room.Encounter.Event)
			g.CompleteRoom(roomID)
		}
	}
	return initialLog, nil
}

// Updates game state
func (g *Game) CompleteRoom(roomID string) {
	room := models.RoomByID(g.Floors, roomID)
	if room == nil {
		return
	}

	if room.IsCompleted {
		return
	}

	room.IsCompleted = true

	fi, _, err := models.GetFloorRoomIdx(roomID)
	if err != nil {
		log.Println(err)
	}

	for ri := range g.Floors[fi].Rooms {
		if g.Floors[fi].Rooms[ri].Id != roomID {
			g.Floors[fi].Rooms[ri].CanEnter = false
		}
	}

	for _, nextID := range room.NextRoomIDs {
		if next := models.RoomByID(g.Floors, nextID); next != nil {
			next.CanEnter = true
		}
	}

	g.Floors[fi].IsCompleted = true
}

// Have an event affect player
func (g *Game) applyEvent(e *models.Event) {
	h := &g.Player
	for stat, delta := range e.StatsAffected {
		switch stat {
		case models.HealthStat:
			prevMax := h.MaxHP()
			h.LevelStats.Health += delta
			newMax := h.MaxHP()
			if delta > 0 {
				h.CurrentHP += newMax - prevMax
			}
			h.CurrentHP = min(max(h.CurrentHP, 1), newMax)
		case models.ManaStat:
			prevMax := h.MaxMana()
			h.LevelStats.Mana += delta
			newMax := h.MaxMana()
			if delta > 0 {
				h.CurrentMana += newMax - prevMax
			}
			h.CurrentMana = min(max(h.CurrentMana, 0), newMax)
		case models.AttackStat:
			h.LevelStats.Attack = max(0, h.LevelStats.Attack+delta)
		case models.DefenseStat:
			h.LevelStats.Defense = max(0, h.LevelStats.Defense+delta)
		case models.MagicStat:
			h.LevelStats.Magic = max(0, h.LevelStats.Magic+delta)
		}
	}

	e.Applied = true
}

// gets a pool of last defeated monster's moves and has the player learn one of them at random
func (g *Game) learnFromMonster() *models.LearnedMove {
	var pool []string
	monster := g.CurrentRoom().Encounter.Monster
	for _, m := range monster.Moves {
		if _, ok := g.AllMoves[m.MoveID]; !ok {
			continue
		}
		pool = append(pool, m.MoveID)
	}

	if len(pool) == 0 {
		return nil
	}

	chosen := pool[mrand.Intn(len(pool))]
	learned := g.Player.LearnMove(chosen, monster.Level)
	return &learned
}

// adds random item to the players random pool
func (g *Game) lootMonster() *models.Item {
	monster := g.CurrentRoom().Encounter.Monster
	var pool []models.Item
	for _, item_id := range monster.ItemPool {
		if item, ok := g.Config.Items[item_id]; ok {
			pool = append(pool, item)
		}
	}
	if len(pool) == 0 {
		return nil
	}

	// shuffle possible items so the chances are random
	// if it weren't suffled, the chance of getting the last item in the list would be smaller than it should be
	mrand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	for i := range pool {
		if mrand.Intn(100) < pool[i].DropRate {
			g.Player.ItemPool = append(g.Player.ItemPool, pool[i].Id)
			return &pool[i]
		}
	}

	return nil
}

// EquipMove validates and equips a move for the hero.
func (g *Game) EquipMove(moveID string, maxSlots int) error {
	hero := &g.Player

	if len(hero.EquippedMoves) >= maxSlots {
		return fmt.Errorf("equipped moves at maximum")
	}

	if slices.Contains(hero.EquippedMoves, moveID) {
		return fmt.Errorf("move already equipped")
	}

	if hero.GetMoveLevel(moveID) == 0 {
		return fmt.Errorf("move not learned")
	}

	hero.EquippedMoves = append(hero.EquippedMoves, moveID)
	return nil
}

// UnequipMove removes a move from the hero's equipped list.
func (g *Game) UnequipMove(moveID string) error {
	hero := &g.Player
	if len(hero.EquippedMoves) <= 1 {
		return fmt.Errorf("must keep at least one move equipped")
	}

	updated := make([]string, 0, len(hero.EquippedMoves)-1)
	for _, id := range hero.EquippedMoves {
		if id != moveID {
			updated = append(updated, id)
		}
	}
	hero.EquippedMoves = updated
	return nil
}

// Utility
// FindInItemPool returns the item with itemID from the player's pool, if present.
func (g *Game) GetItemFromPlayerPool(itemID string) (models.Item, bool) {
	if slices.Contains(g.Player.ItemPool, itemID) {
		item, ok := g.Config.Items[itemID]
		return item, ok
	}

	return models.Item{}, false
}

// Utility
// RemoveFromItemPool removes one occurrence of itemID from the player's pool.
func (g *Game) RemoveFromItemPool(itemID string) {
	for i, id := range g.Player.ItemPool {
		if id == itemID {
			g.Player.ItemPool = slices.Delete(g.Player.ItemPool, i, i+1)
			return
		}
	}
}

// EquipItemFromPool equips an item from the player's inventory.
func (g *Game) EquipItemFromPool(itemID string) error {
	item, ok := g.GetItemFromPlayerPool(itemID)
	if !ok {
		return fmt.Errorf("item not in inventory")
	}
	if err := g.Player.EquipItem(item); err != nil {
		return err
	}
	g.RemoveFromItemPool(itemID)
	return nil
}

// UseItemFromPool applies a consumable from the player's inventory.
func (g *Game) UseItemFromPool(itemID string) error {
	item, ok := g.GetItemFromPlayerPool(itemID)
	if !ok {
		return fmt.Errorf("item not in inventory")
	}
	if err := g.Player.ApplyConsumableItem(item); err != nil {
		return err
	}
	g.RemoveFromItemPool(itemID)
	return nil
}
