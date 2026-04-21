package models

type Hero struct {
	Entity
	Name          string
	Level         int
	CurrentXP     int
	LevelBonuses  Stats
	LearnedMoves  []Move
	EquippedMoves []Move
}
