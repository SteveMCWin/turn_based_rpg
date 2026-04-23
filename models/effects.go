package models

type EffectType string

const (
	Permanent EffectType = "permanent"
	Temporary EffectType = "temporary"
)

type EffectTarget string

const (
	TargetSelf     EffectTarget = "self"
	TargetOpponeng EffectTarget = "opponent"
)

// Couldn't come up with a better naming, but the effect is
// the effect an attack will apply, the definition
type Effect struct {
	Type            EffectType   `json:"type"`
	StatAffected    StatType     `json:"stat_affected"`
	Delta           int          `json:"delta"`
	Duration        int          `json:"duration"`
	Target          EffectTarget `json:"target"`
	CostAmount      int          `json:"cost_amount,omitempty"`
	CostType        StatType     `json:"cost_type,omitempty"`
	ActivationTimer int          `json:"turns_to_activate"`
}

// This is used to identify the status of a character
// So an attack has a 'burn' effect, and when it
// hits an opponent, the oponent gets a StatusEffect 'burn'
type StatusEffect struct {
	Effect
	TurnsRemaining  int
	TurnsToActivate int
}
