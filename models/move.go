package models

// Used for calculating damage (magical bypasses defense)
type MoveType string

const (
	Physical MoveType = "physical"
	Magical  MoveType = "magical"
)

// Used for the enemy AI
type MoveIntent string

const (
	IntentDamage MoveIntent = "damage"
	IntentHeal   MoveIntent = "heal"
	IntentBuff   MoveIntent = "buff"
	IntentDebuff MoveIntent = "debuff"
)

// The base move
// Id, name and description are self explanatory
// Move type, move intent are explained above
// Effects are the effects that the move will apply to an entity
// Base value is one part of the equation of how much impact a move will have
// The other part is scaling stat and the entities stats * ScalingStatFactor.
// If the move scales with magic, more magic -> stronger move
// Cost amount is how much mana the move uses up
type MoveDefinition struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	BaseValue         int        `json:"base_value"`
	MoveType          MoveType   `json:"move_type"`
	Intent            MoveIntent `json:"primary"`
	Effects           []*Effect  `json:"effects,omitempty"`
	ScalingStat       StatType   `json:"scaling_stat"`
	ScalingStatFactor float32    `json:"scaling_stat_amount"`
	CostAmount        int        `json:"cost_amount,omitempty"`
}

// Learned move is a combination of a base move and a level of the move
// Since many entities share moves, levels differ, but the base is the same
type LearnedMove struct {
	MoveID string `json:"move_id"`
	Level  int    `json:"level"`
}
