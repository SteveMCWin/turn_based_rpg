package models

import (
	"fmt"
	"slices"
)

// Entity represents the base template for the player and monsters alike
// level is calculated from CurrentXP once AddXP is called
// since the stats themself are just the base stats, we keep another stats field (LevelStats)
// that accumulates the stat growth from leveling
// StatScaling is how big of a % increase a level-up will bring to the base stats
// Status effects are what effects are applied on the entity
// Item pool is all the items the monster may have/drop. For a player it's starting items
// Equipped items actually affect the entities stats
// EnvironmentEffects are different from status effects since each entity has baked into itself
// how an environment affects it, so it's a map of ("environment_id" : effeft it applies)
// e.g. the skeleton may feel at home in the crypt and get a buff, but the knight may find it a bit spooky and lose defense
type Entity struct {
	Stats
	Level         int              `json:"level"`
	CurrentXP     int              `json:"current_xp"`
	CurrentHP     int              `json:"current_hp"`
	CurrentMana   int              `json:"current_mana"`
	LevelStats    Stats            `json:"level_stats"`
	StatScaling   StatScaleFactors `json:"stat_scaling"`
	StatusEffects []StatusEffect   `json:"status_effects"`
	ItemPool      []string         `json:"item_pool"`
	EquippedItems []Item           `json:"equipment"`

	EnvironmentEffects map[string]Effect `json:"env_effects"`
}

func (e *Entity) IsAlive() bool {
	return e.CurrentHP > 0
}

// Slightly more efficient and conveniet subset of EffectiveStats
func (e *Entity) MaxHP() int {
	hp := e.Health + e.LevelStats.Health
	for _, item := range e.EquippedItems {
		if item.Type != Consumable {
			hp += item.StatsAffected[HealthStat]
		}
	}
	return hp
}

// Slightly more efficient and conveniet subset of EffectiveStats
func (e *Entity) MaxMana() int {
	mana := e.Stats.Mana + e.LevelStats.Mana
	for _, item := range e.EquippedItems {
		if item.Type != Consumable {
			mana += item.StatsAffected[ManaStat]
		}
	}
	return mana
}

// Since e.Stats is just the base, meant for level 1 characters,
// we need to calculate stats based on modifiers like character level
// and [de]buffs and items
func (e *Entity) EffectiveStats() Stats {
	s := e.Stats.Add(e.LevelStats)
	for _, se := range e.StatusEffects {
		switch se.Type {
		case StatModifier:
			if se.TurnsToActivate <= 0 {
				s = s.ApplyEffect(se.Effect)
			}
		}
	}

	for _, item := range e.EquippedItems {
		if item.Type == Consumable {
			continue
		}

		for stat, delta := range item.StatsAffected {
			s = s.AddToStat(stat, delta)
		}
	}

	return s
}

// append item to entities equipped items if possible
func (e *Entity) EquipItem(item Item) error {
	if item.Type == Consumable {
		return fmt.Errorf("Cannot equip consumable item")
	}

	for _, equipped := range e.EquippedItems {
		if equipped.Type == item.Type {
			return fmt.Errorf("Cannot have two items of the same type equipped")
		}
	}

	e.EquippedItems = append(e.EquippedItems, item)
	return nil
}

func (e *Entity) UnequipItem(item_id string) {
	for i, item := range e.EquippedItems {
		if item.Id == item_id {
			e.EquippedItems = slices.Delete(e.EquippedItems, i, i+1)
		}
	}
}

// Note that consubales at the moment only apply to current hp and mana
func (e *Entity) ApplyConsumableItem(item Item) error {
	if item.Type != Consumable {
		return fmt.Errorf("Cannot consume non-consumable item")
	}

	for stat, delta := range item.StatsAffected {
		switch stat {
		case HealthStat:
			e.CurrentHP = min(e.MaxHP(), e.CurrentHP+delta)
		case ManaStat:
			e.CurrentMana = min(e.MaxMana(), e.CurrentMana+delta)
		}
	}

	return nil
}

// Scale entites level bonuses (LevelStats) based on the stat scaling of that entity and their level
func (e *Entity) SetToLevel(new_level int) {
	current_lvl := e.Level

	prevHpPercent := float32(e.CurrentHP) / float32(e.MaxHP())
	prevManaPercent := float32(e.CurrentMana) / float32(e.MaxMana())

	e.LevelStats.Health += int(float32(e.Stats.Health)*e.StatScaling.HealthScaling) * (new_level - current_lvl)
	e.LevelStats.Mana += int(float32(e.Stats.Mana)*e.StatScaling.ManaScaling) * (new_level - current_lvl)
	e.LevelStats.Attack += int(float32(e.Stats.Attack)*e.StatScaling.AttackScaling) * (new_level - current_lvl)
	e.LevelStats.Defense += int(float32(e.Stats.Defense)*e.StatScaling.DefenseScaling) * (new_level - current_lvl)
	e.LevelStats.Magic += int(float32(e.Stats.Magic)*e.StatScaling.MagicScaling) * (new_level - current_lvl)

	// give player 20% max hp and mana heal as reward
	prevHpPercent = min(1.0, prevHpPercent+0.2)
	prevManaPercent = min(1.0, prevManaPercent+0.2)

	e.CurrentHP = int(float32(e.MaxHP()) * prevHpPercent)
	e.CurrentMana = int(float32(e.MaxMana()) * prevManaPercent)

	e.Level = new_level
}

// used when the player can choose which stats to upgrade on level up
func (e *Entity) LevelUpStat(stat StatType, amount int) {
	prevMaxHP := e.MaxHP()
	prevMaxMana := e.MaxMana()

	switch stat {
	case HealthStat:
		e.LevelStats.Health += int(float32(e.Stats.Health)*e.StatScaling.HealthScaling) * amount
	case ManaStat:
		e.LevelStats.Mana += int(float32(e.Stats.Mana)*e.StatScaling.ManaScaling) * amount
	case AttackStat:
		e.LevelStats.Attack += int(float32(e.Stats.Attack)*e.StatScaling.AttackScaling) * amount
	case DefenseStat:
		e.LevelStats.Defense += int(float32(e.Stats.Defense)*e.StatScaling.DefenseScaling) * amount
	default:
		e.LevelStats.Magic += int(float32(e.Stats.Magic)*e.StatScaling.MagicScaling) * amount
	}

	e.CurrentHP += e.MaxHP() - prevMaxHP
	e.CurrentMana += e.MaxMana() - prevMaxMana
}

// decrease time to activate or turns remaining of each of the entities effects and remove them if expired
// also deal damage if DoT effect
func (e *Entity) TickStatusEffects() {
	effects_to_remove := []int{}
	for i, se := range e.StatusEffects {

		if e.StatusEffects[i].TurnsToActivate > 0 {
			e.StatusEffects[i].TurnsToActivate--
		}

		if e.StatusEffects[i].TurnsToActivate <= 0 && e.StatusEffects[i].TurnsRemaining > 0 {
			switch se.Type {
			case DamageOverTime:
				e.CurrentHP = max(0, e.CurrentHP-se.Effect.BaseDelta)
			}

			if !e.StatusEffects[i].IsEnvironmental {
				e.StatusEffects[i].TurnsRemaining--
			}

			if e.StatusEffects[i].TurnsRemaining <= 0 {
				effects_to_remove = append(effects_to_remove, i)
			}
		}
	}

	for i := len(effects_to_remove) - 1; i >= 0; i-- {
		idx := effects_to_remove[i]
		e.StatusEffects = slices.Delete(e.StatusEffects, idx, idx+1)
	}
}

func (e *Entity) ClearStatusEffects() {
	e.StatusEffects = make([]StatusEffect, 0)
}

// Called after battle. XP is linear with the level of the enemy beaten
// thresholds is a slice (which is a dynamic array in golang) of how much xp is needed for leveling up
// returns how many levels the player got based on the level thresholds
func (e *Entity) AddXP(amount int, thresholds []int) int {
	e.CurrentXP += amount
	levels_gained := 0
	for e.Level < len(thresholds) && e.CurrentXP >= thresholds[e.Level] {
		e.SetToLevel(e.Level + 1)
		levels_gained++
	}
	return levels_gained
}
