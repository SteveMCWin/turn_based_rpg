package game

import (
	crand "crypto/rand"
	mrand "math/rand"
	"fmt"
	"log"
	"slices"

	"tbrpg/models"
)

type Game struct {
	ID            string         `json:"id"`
	Settings      GameSettings   `json:"settings"`
	Player        models.Hero    `json:"player"`
	Floors        []models.Floor `json:"floors"`
	BattleLog     []string       `json:"battle_log,omitempty"`
	IsInBattle    bool           `json:"in_battle"`
	CurrentRoomID string         `json:"current_room_id,omitempty"`

	AllMoves map[string]models.MoveDefinition `json:"moves,omitempty"`
}

func NewGame(config *GameConfig) *Game {

	events := slices.Clone(config.EventTemplates)
	monsters := slices.Clone(config.MonsterTemplates)
	for i := range monsters {
		monsters[i].Init()
	}

	g := Game{}

	g.ID = newUUID()
	g.Settings = config.Settings
	g.Player = config.HeroTemplate
	g.Player.Init()
	g.Floors = models.GenerateFloors(len(monsters), config.Settings.MaxRoomsPerLevel)
	models.FillFloorEncounters(g.Floors, monsters, events)

	return &g
}

// for testing purposes, later on the id will be assigned by the database
func newUUID() string {
	var b [16]byte
	crand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
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
	if room.IsCompleted {
		return fmt.Errorf("room %s is already completed", roomID)
	}

	switch room.Encounter.Kind {
	case models.EncounterKindMonster:
		g.BattleLog = nil
		if room.IsCompleted {

			g.BattleLog = append(g.BattleLog, "You applied black magic to revive an already defeated foe. They are as hostile as you remember them to be.")

			if mrand.Int()%100 <= g.Settings.PercentChanceMonsterLevelsUp {
				room.Encounter.Monster.LevelUp()
				g.BattleLog = append(g.BattleLog, "However, upon reviving the monster for another duel, something went wrong in the ritual, and the monster is now permanently stronger!")
			}
		}

		room.Encounter.Monster.ResetForBattle()
		g.CurrentRoomID = roomID
		g.IsInBattle = true

	case models.EncounterKindEvent:
		if room.Encounter.Event != nil && !room.Encounter.Event.Applied {
			g.applyEvent(room.Encounter.Event)
		}
		g.CompleteRoom(roomID)
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
