package game

import (
	"tbrpg/models"
	"time"
)

type Game struct {
	ID     string
	Config GameConfig
	Player models.Hero
	Floors []models.Floor
}

func NewGame(conf GameConfig) Game {

	g := Game{}

	g.ID = time.Now().Format("2006-01-02 15:04:05")
	g.Config = conf
	g.Player = models.Hero{}
	g.Player.Init()
	g.Floors = models.GenerateFloors(len(conf.MonsterTemplates), conf.Settings.MaxRoomsPerLevel)
	models.FillFloorEncounters(g.Floors, conf.MonsterTemplates, conf.EventTemplates)

	return g
}
