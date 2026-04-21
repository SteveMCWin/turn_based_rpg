package models

import "log"

// Stats shared by all entities
type Stats struct {
	Health  int
	Attack  int
	Defense int
	Magic   int
}

// Percentage of increase for each stat after level gain
// i.e. if HealthScaling is 0.2, health increases by 20% on level-up
type StatScaleFactors struct {
	HealthScaling  float32
	AttackScaling  float32
	DefenseScaling float32
	MagicScaling   float32
}

// Avoids string misspells
type StatType string

const (
	HealthStat  StatType = "health"
	AttackStat  StatType = "attack"
	DefenseStat StatType = "defense"
	MagicStat   StatType = "magic"
)

func (s Stats) ScaleStats(scaleFactor StatScaleFactors) Stats {
	new_stats := Stats {
		Health: s.Health + s.Health * int(scaleFactor.HealthScaling),
		Attack: s.Attack + s.Attack * int(scaleFactor.AttackScaling),
		Defense: s.Defense + s.Defense * int(scaleFactor.DefenseScaling),
		Magic: s.Magic + s.Magic * int(scaleFactor.MagicScaling),
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
	default:
		log.Println("Urmmm")
	}

	return s
}
