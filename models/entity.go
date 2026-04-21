package models

import "slices"

const (
	MaxMoveLevel     = 5
	MaxEquippedMoves = 4
)

type Entity struct {
	Stats
	CurrentHP int
	StatusEffects []StatusEffect
}

func (e *Entity) IsAlive() bool {
	return e.CurrentHP > 0
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
