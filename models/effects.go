package models

type EffectType string

const (
	StatModifier EffectType = "stat_mod"
	DamageOverTime EffectType = "dot"
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
	ActivationDelay int          `json:"activation_delay"`
}

// This is used to identify the status of a character
// So an attack has a 'burn' effect, and when it
// hits an opponent, the oponent gets a StatusEffect 'burn'
type StatusEffect struct {
	Effect
	TurnsRemaining  int
	TurnsToActivate int
}
