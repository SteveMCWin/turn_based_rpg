package game

import (
	"fmt"
	"log"
	mrand "math/rand"
	"slices"

	"tbrpg/models"
)

type Game struct {
	ID               int                              `json:"id"`
	Settings         GameSettings       `json:"settings"`
	Player           models.Hero        `json:"player"`
	Floors           []models.Floor     `json:"floors"`
	BattleLog        []string           `json:"battle_log,omitempty"`
	IsInBattle       bool               `json:"in_battle"`
	CurrentRoomID    string             `json:"current_room_id,omitempty"`
	LastBattleResult *BattleResult      `json:"last_battle_result"`
	PendingLevelUp   *PendingAllocation `json:"pending_level_up"`

	AllMoves map[string]models.MoveDefinition `json:"moves,omitempty"`
}

type PendingAllocation struct {
	ManualPoints int `json:"manual_points"`
	RandomPoints int `json:"random_points"`
}

func NewGame(config *GameConfig, hero models.Hero) *Game {
	events := slices.Clone(config.EventTemplates)

	monsters := slices.Clone(config.MonsterTemplates)
	for i := range monsters {
		monsters[i].Init()
	}

	g := Game{
		Settings: config.Settings,
		Player:   hero,
		Floors:   models.GenerateFloors(len(monsters), config.Settings.MaxRoomsPerLevel),
		AllMoves: config.Moves,
	}

	g.Player.Init()
	models.FillFloorEncounters(g.Floors, monsters, events)

	return &g
}

func (g *Game) RoomByID(id string) *models.Room {
	for fi := range g.Floors {
		for ri := range g.Floors[fi].Rooms {
			if g.Floors[fi].Rooms[ri].ID == id {
				return &g.Floors[fi].Rooms[ri]
			}
		}
	}
	return nil
}

func (g *Game) CurrentRoom() *models.Room {
	if g.CurrentRoomID == "" {
		return nil
	}

	return g.RoomByID(g.CurrentRoomID)
}

func (g *Game) EnterRoom(roomID string) error {
	room := g.RoomByID(roomID)
	if room == nil {
		return fmt.Errorf("room not found: %s", roomID)
	}
	if !room.CanEnter {
		return fmt.Errorf("room %s is not accessible", roomID)
	}

	switch room.Encounter.Kind {
	case models.EncounterKindMonster:
		g.BattleLog = nil
		if room.IsCompleted {

			g.BattleLog = append(g.BattleLog, "You applied black magic to revive an already defeated foe. They are as hostile as you remember them to be.")

			if mrand.Int()%100 <= g.Settings.PercentChanceMonsterLevelsUp {
				room.Encounter.Monster.SetToLevel(room.Encounter.Monster.Level+1)
				g.BattleLog = append(g.BattleLog, "However, upon reviving the monster for another duel, something went wrong in the ritual, and the monster is now permanently stronger!")
			}
		}

		room.Encounter.Monster.ResetForBattle()
		g.CurrentRoomID = roomID
		g.IsInBattle = true

	case models.EncounterKindEvent:
		if room.Encounter.Event != nil && !room.Encounter.Event.Applied {
			g.applyEvent(room.Encounter.Event)
			g.CompleteRoom(roomID)
		}
		// g.CompleteRoom(roomID)
	}
	return nil
}

func (g *Game) CompleteRoom(roomID string) {
	room := g.RoomByID(roomID)
	if room == nil {
		return
	}
	room.IsCompleted = true

	fi, _, err := models.GetFloorIdx(roomID)
	if err != nil {
		log.Println(err)
	}

	for ri := range g.Floors[fi].Rooms {
		if g.Floors[fi].Rooms[ri].ID != roomID {
			g.Floors[fi].Rooms[ri].CanEnter = false
		}
	}

	for _, nextID := range room.NextRoomIDs {
		if next := g.RoomByID(nextID); next != nil {
			next.CanEnter = true
		}
	}

	g.Floors[fi].IsCompleted = true
}

func (g *Game) applyEvent(e *models.Event) {
	h := &g.Player
	switch e.StatAffected {
	case models.HealthStat:
		prevMax := h.MaxHP()
		h.LevelBonuses.Health += e.Delta
		newMax := h.MaxHP()
		if e.Delta > 0 {
			h.CurrentHP += newMax - prevMax
		}
		h.CurrentHP = min(max(h.CurrentHP, 1), newMax)
	case models.AttackStat:
		h.LevelBonuses.Attack = max(0, h.LevelBonuses.Attack+e.Delta)
	case models.DefenseStat:
		h.LevelBonuses.Defense = max(0, h.LevelBonuses.Defense+e.Delta)
	case models.MagicStat:
		h.LevelBonuses.Magic = max(0, h.LevelBonuses.Magic+e.Delta)
	case models.ManaStat:
		prevMax := h.MaxMana()
		h.LevelBonuses.Mana += e.Delta
		newMax := h.MaxMana()
		if e.Delta > 0 {
			h.CurrentMana += newMax - prevMax
		}
		h.CurrentMana = min(max(h.CurrentMana, 0), newMax)
	}
	e.Applied = true
}

func (g *Game) learnRandomMove() *models.LearnedMove {
	var pool []string
	monster := g.CurrentRoom().Encounter.Monster
	for _, m := range monster.Moves {
		if _, ok := g.AllMoves[m.MoveID]; !ok {
			continue
		}
		if g.Player.GetMoveLevel(m.MoveID) < g.Settings.MaxMoveLevel {
			pool = append(pool, m.MoveID)
		}
	}

	if len(pool) == 0 {
		return nil
	}

	chosen := pool[mrand.Intn(len(pool))]
	learned := g.Player.LearnMove(chosen, monster.Level)
	return &learned
}
