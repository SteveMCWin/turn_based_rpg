package models

import "slices"

// Entity represents the base template for the player and monsters alike
// level is calculated from CurrentXP once AddXP is called
// since the stats themself are just the base stats, we been another stats field (LevelBonuses)
// that is a modifier with which we keep track of the actual current stats
// StatScaling is how big of a % increase a level-up will bring to the base stats
type Entity struct {
	Stats
	Level         int              `json:"level"`
	CurrentXP     int              `json:"current_xp"`
	LevelBonuses  Stats            `json:"level_bonuses"`
	StatScaling   StatScaleFactors `json:"stat_scaling"`
	CurrentHP     int              `json:"current_hp"`
	CurrentMana   int              `json:"current_mana"`
	StatusEffects []StatusEffect   `json:"status_effects"`
}

func (e *Entity) IsAlive() bool {
	return e.CurrentHP > 0
}

func (e *Entity) MaxHP() int {
	return e.Health + e.LevelBonuses.Health
}

func (e *Entity) MaxMana() int {
	return e.Stats.Mana + e.LevelBonuses.Mana
}

// Since e.Stats is just the base, meant for level 1 characters,
// we need to calculate stats based on modifiers like character level
// and debuffs or buffs
func (e *Entity) EffectiveStats() Stats {
	s := e.Stats.Add(e.LevelBonuses)
	for _, se := range e.StatusEffects {
		switch se.Type {
		case StatModifier:
			if se.TurnsToActivate <= 0 {
				s = s.ApplyEffect(se.Effect)
			}
		}
	}
	return s
}

// Increases entites level bonuses based on the stat scaling of that entity
func (e *Entity) LevelUpStats() {
	prevMaxHP := e.MaxHP()
	prevMaxMana := e.MaxMana()

	e.LevelBonuses.Health += int(float32(e.Stats.Health) * e.StatScaling.HealthScaling)
	e.LevelBonuses.Attack += int(float32(e.Stats.Attack) * e.StatScaling.AttackScaling)
	e.LevelBonuses.Defense += int(float32(e.Stats.Defense) * e.StatScaling.DefenseScaling)
	e.LevelBonuses.Magic += int(float32(e.Stats.Magic) * e.StatScaling.MagicScaling)
	e.LevelBonuses.Mana += int(float32(e.Stats.Mana) * e.StatScaling.ManaScaling)

	// add health and mana on level up as a reward :^D
	e.CurrentHP += e.MaxHP() - prevMaxHP
	e.CurrentMana += e.MaxMana() - prevMaxMana
}


func (e *Entity) TickStatusEffects() {
	effects_to_remove := []int{}
	for i, se := range e.StatusEffects {

		if e.StatusEffects[i].TurnsToActivate > 0 {
			e.StatusEffects[i].TurnsToActivate--
		}

		if e.StatusEffects[i].TurnsToActivate <= 0 && e.StatusEffects[i].TurnsRemaining > 0 {
			switch se.Type {
			case DamageOverTime:
				e.CurrentHP = max(0, e.CurrentHP - se.Effect.Delta)
			}

			e.StatusEffects[i].TurnsRemaining--
			if e.StatusEffects[i].TurnsRemaining <= 0 {
				effects_to_remove = append(effects_to_remove, i)
			}
		}
	}

	for i := len(effects_to_remove)-1; i >= 0; i-- {
		idx := effects_to_remove[i]
		e.StatusEffects = slices.Delete(e.StatusEffects, idx, idx+1)
	}
}

func (e *Entity) ClearStatusEffects() {
	e.StatusEffects = make([]StatusEffect, 0)
}

func (e *Entity) AddXP(amount int, thresholds []int) int {
	e.CurrentXP += amount
	levels_gained := 0
	for e.Level < len(thresholds) && e.CurrentXP >= thresholds[e.Level] {
		e.LevelUp()
		levels_gained++
	}
	return levels_gained
}

func (e *Entity) LevelUp() {
	e.Level += 1
	e.LevelUpStats()
}
