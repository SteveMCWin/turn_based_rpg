package models

import "log"

// Stats shared by all entities
type Stats struct {
	Health  int `json:"health"`
	Mana    int `json:"mana"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	Magic   int `json:"magic"`
	// resilience against effects? dodge? out of scope...
}

// Percentage of increase for each stat after level gain
// i.e. if HealthScaling is 0.2, health increases by 20% on level-up
type StatScaleFactors struct {
	HealthScaling  float32 `json:"health_scaling"`
	ManaScaling    float32 `json:"mana_scaling"`
	AttackScaling  float32 `json:"attack_scaling"`
	DefenseScaling float32 `json:"defense_scaling"`
	MagicScaling   float32 `json:"magic_scaling"`
}

// Avoids string misspells
type StatType string

const (
	HealthStat  StatType = "health"
	ManaStat    StatType = "mana"
	AttackStat  StatType = "attack"
	DefenseStat StatType = "defense"
	MagicStat   StatType = "magic"
)

// Just combine stats
// used mostly for calculating current 'real' stats
func (s Stats) Add(other Stats) Stats {
	return Stats{
		Health:  s.Health + other.Health,
		Mana:    s.Mana + other.Mana,
		Attack:  s.Attack + other.Attack,
		Defense: s.Defense + other.Defense,
		Magic:   s.Magic + other.Magic,
	}
}

// Increases stats by a percentage defined for each stat in scaleFactor
func (s Stats) ScaleStats(scaleFactor StatScaleFactors) Stats {
	new_stats := Stats{
		Health:  s.Health + int(float32(s.Health)*scaleFactor.HealthScaling),
		Mana:   s.Mana + int(float32(s.Mana)*scaleFactor.ManaScaling),
		Attack:  s.Attack + int(float32(s.Attack)*scaleFactor.AttackScaling),
		Defense: s.Defense + int(float32(s.Defense)*scaleFactor.DefenseScaling),
		Magic:   s.Magic + int(float32(s.Magic)*scaleFactor.MagicScaling),
	}

	return new_stats
}

// Affect stats based on effect type
func (s Stats) ApplyEffect(e Effect) Stats {
	switch e.StatAffected {
	case HealthStat:
		s.Health += e.BaseDelta
		s.Health = max(s.Health, 0)
	case ManaStat:
		s.Mana += e.BaseDelta
		s.Mana = max(s.Mana, 0)
	case AttackStat:
		s.Attack += e.BaseDelta
		s.Attack = max(s.Attack, 0)
	case DefenseStat:
		s.Defense += e.BaseDelta
		s.Defense = max(s.Defense, 0)
	case MagicStat:
		s.Magic += e.BaseDelta
		s.Magic = max(s.Magic, 0)
	default:
		log.Println("Urmmm")
	}

	return s
}

// Apply delta to a stat based on stat type, used for item buffs
func (s Stats) AddToStat(stat StatType, delta int) Stats {
	res := s

	switch stat {
	case HealthStat:
		res.Health += delta
	case ManaStat:
		res.Mana += delta
	case AttackStat:
		res.Attack += delta
	case DefenseStat:
		res.Defense += delta
	case MagicStat:
		res.Magic += delta
	}

	return res
}
