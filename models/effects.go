package models

type EffectType string

// Stat modifiers are temporary
// Damage over time modifies current health permanently
const (
	StatModifier   EffectType = "stat_mod"
	DamageOverTime EffectType = "dot"
)

type EffectTarget string

const (
	TargetSelf     EffectTarget = "self"
	TargetOpponent EffectTarget = "opponent"
)

// Couldn't come up with a better naming, but the effect is
// the effect an attack will apply, the definition
// Type, stat affected, target, duration and activation delay shoul dbe self explanatory
// if there is a scale factor, it multiplies the stat the move that has this effect scales with
// if there is no scale factor a constant base delta is used
// not a big fan of how I did it either
type Effect struct {
	Type            EffectType   `json:"type"`
	StatAffected    StatType     `json:"stat_affected"`
	BaseDelta       int          `json:"delta"`
	ScaleFactor     float32      `json:"scale_factor,omitempty"`
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
	IsEnvironmental bool
}
