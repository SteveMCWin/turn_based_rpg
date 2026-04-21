package models

type EffectType string

const (
	Lasting       EffectType = "lasting"
	StatModifier  EffectType = "stat_modifier"
	OneTimeEffect EffectType = "one_time"
)

type EffectTarget string

const (
	TargetSelf     EffectTarget = "self"
	TargetOpponeng EffectTarget = "opponent"
)

// Couldn't come up with a better naming, but the effect is
// the effect an attack will apply, the definition
type Effect struct {
	Type         EffectType
	StatAffected StatType
	Delta        int
	Duration     int
	Target       EffectTarget
	CostAmount   int
	CostType     StatType
}

// This is used to identify the status of a character
// So an attack has a 'burn' effect, and when it
// hits an opponent, the oponent gets a status effect 'burn'
type StatusEffect struct {
	Effect
	TurnsRemaining int
}
