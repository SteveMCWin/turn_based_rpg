package models

import "slices"

const (
	MaxMoveLevel     = 5
	MaxEquippedMoves = 4
)

// Entity represents the base for the player and monsters alike
// level is calculated from CurrentXP once AddXP is called
// since the stats themself are just the base stats, we been another stats field (LevelBonuses)
// that is a modifier with which we keep track of the actual current stats
// StatScaling is how big of a % increase a level-up will bring to the base stats

type Entity struct {
	Stats
	Level         int
	CurrentXP     int
	LevelBonuses  Stats
	StatScaling   StatScaleFactors
	CurrentHP     int
	StatusEffects []StatusEffect
}

func (e *Entity) IsAlive() bool {
	return e.CurrentHP > 0
}

func (e *Entity) MaxHP() int {
	return e.Health + e.LevelBonuses.Health
}

func (e *Entity) ApplyStatusEffects() {
	effects_to_remove := []int{}
	for i, se := range e.StatusEffects {
		switch se.Type {
		case OneTimeEffect:
			e.Stats = e.ApplyEffect(se.Effect)
			effects_to_remove = append(effects_to_remove, i)

		case Lasting:
			e.Stats = e.ApplyEffect(se.Effect)
			se.Duration -= 1

			if se.Duration == 0 {
				effects_to_remove = append(effects_to_remove, i)
			}

		case StatModifier:
			// TODO
		}
	}

	for _, idx := range effects_to_remove {
		e.StatusEffects = slices.Delete(e.StatusEffects, idx, idx+1)
	}
}

func (e *Entity) ClearStatusEffects() {
	e.StatusEffects = make([]StatusEffect, 0)
}

func (e *Entity) AddXP(amount int, thresholds []int, statPerLevel Stats) int {
	e.CurrentXP += amount
	gained := 0
	for e.Level < len(thresholds) && e.CurrentXP >= thresholds[e.Level] {
		e.LevelUp()
		e.LevelBonuses = e.LevelBonuses.Add(statPerLevel)
		gained++
	}
	return gained
}

func (e *Entity) LevelUp() {
	e.Level += 1
}
