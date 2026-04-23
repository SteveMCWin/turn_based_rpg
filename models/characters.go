package models

// The entity field is the base entity struct with stats and status effects
// learned moves are ids of all the moves a player can equip at the moment
// equipped moves is self explanatory :^)
type Hero struct {
	Entity
	Name          string
	LearnedMoves  []LearnedMove
	EquippedMoves []string
}

func (h *Hero) Init() {
	h.CurrentHP = h.Stats.Health
	h.CurrentMana = h.Stats.Mana
	h.StatusEffects = make([]StatusEffect, 0)
	if len(h.LearnedMoves) > 0 {
		h.EquippedMoves = make([]string, len(h.LearnedMoves))
		for i, m := range h.LearnedMoves {
			h.EquippedMoves[i] = m.MoveID
		}
	}
}

func (h *Hero) LearnMove(m LearnedMove) {
	h.LearnedMoves = append(h.LearnedMoves, m)
}

// Monster represents an enemy in the gauntlet.
type Monster struct {
	Entity
	ID         string
	Name       string
	Moves      []LearnedMove
	IsDefeated bool
}

func (m *Monster) ResetForBattle() {
	m.CurrentHP = m.MaxHP()
	m.ClearStatusEffects()
}
