package models

type MoveType string

const (
	Physical MoveType = "physical"
	Magical  MoveType = "magical"
)

type Move struct {
	Name         string
	Description  string
	MoveType     MoveType
	Effect       Effect
	BaseValue    int
	ScalingStat  StatType
	CurrentLevel int
}

