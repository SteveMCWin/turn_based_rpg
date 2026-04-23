package models

// Used for scaling with stats
type MoveType string

const (
	Physical MoveType = "physical"
	Magical  MoveType = "magical"
)

// Used for the enemy AI
type PrimaryType string

const (
	PrimaryDamage PrimaryType = "damage"
	PrimaryHeal   PrimaryType = "heal"
	PrimaryNone   PrimaryType = "none"
)

type MoveDefinition struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	MoveType    MoveType    `json:"move_type"`
	Primary     PrimaryType `json:"primary"`
	Effects     []*Effect   `json:"effect,omitempty"`
	BaseValue   int         `json:"base_value"`
	ScalingStat StatType    `json:"scaling_stat"`
}

type LearnedMove struct {
	MoveID string `json:"move_id"`
	Level  int    `json:"level"`
}
