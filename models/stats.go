package models

import "log"

// Stats shared by all entities
type Stats struct {
	Health  int
	Attack  int
	Defense int
	Magic   int
	Mana    int
	// resilience against effects? dodge?
}

// Percentage of increase for each stat after level gain
// i.e. if HealthScaling is 0.2, health increases by 20% on level-up
type StatScaleFactors struct {
	HealthScaling  float32
	AttackScaling  float32
	DefenseScaling float32
	MagicScaling   float32
	ManaScaling    float32
}

// Avoids string misspells
type StatType string

const (
	HealthStat  StatType = "health"
	AttackStat  StatType = "attack"
	DefenseStat StatType = "defense"
	MagicStat   StatType = "magic"
	ManaStat    StatType = "mana"
)

// Just combine stats
// used mostly for calculating current stats
func (s Stats) Add(other Stats) Stats {
	return Stats{
		Health:  s.Health + other.Health,
		Attack:  s.Attack + other.Attack,
		Defense: s.Defense + other.Defense,
		Magic:   s.Magic + other.Magic,
		Mana:    s.Mana + other.Mana,
	}
}

func (s Stats) ScaleStats(scaleFactor StatScaleFactors) Stats {
	new_stats := Stats{
		Health:  s.Health + int(float32(s.Health)*scaleFactor.HealthScaling),
		Attack:  s.Attack + int(float32(s.Attack)*scaleFactor.AttackScaling),
		Defense: s.Defense + int(float32(s.Defense)*scaleFactor.DefenseScaling),
		Magic:   s.Magic + int(float32(s.Magic)*scaleFactor.MagicScaling),
		Mana:   s.Mana + int(float32(s.Mana)*scaleFactor.ManaScaling),
	}

	return new_stats
}

func (s Stats) ApplyEffect(e Effect) Stats {
	switch e.StatAffected {
	case HealthStat:
		s.Health += e.Delta
		s.Health = max(s.Health, 0)
	case AttackStat:
		s.Attack += e.Delta
		s.Attack = max(s.Attack, 0)
	case DefenseStat:
		s.Defense += e.Delta
		s.Defense = max(s.Defense, 0)
	case MagicStat:
		s.Magic += e.Delta
		s.Magic = max(s.Magic, 0)
	case ManaStat:
		s.Mana += e.Delta
		s.Mana = max(s.Mana, 0)
	default:
		log.Println("Urmmm")
	}

	return s
}
