package game

import (
	"crypto/rand"
	"fmt"
	"slices"

	"tbrpg/models"
)

type Game struct {
	ID     string `json:"id"`
	Settings GameSettings `json:"settings"`
	Player models.Hero `json:"player"`
	Floors []models.Floor `json:"floors"`
}

func NewGame(config *GameConfig) *Game {

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
	models.FillFloorEncounters(g.Floors, monsters, config.EventTemplates)

	return &g
}

func newUUID() string {
	var b [16]byte
	rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
