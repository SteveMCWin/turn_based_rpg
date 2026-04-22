package models

// The entity field is the base entity struct with stats and status effects
// learned moves are all the moves a player can equip at the moment
// equipped moves is self explanatory :^)
type Hero struct {
	Entity
	Name          string
	LearnedMoves  []Move
	EquippedMoves []*Move
}

func (h *Hero) LearnMove(m Move) {
	h.LearnedMoves = append(h.LearnedMoves, m)
}

// Monster represents an enemy in the gauntlet.
type Monster struct {
	Entity
	ID         string
	Name       string
	Moves      []Move
	IsDefeated bool
}

func (m *Monster) ResetForBattle() {
	m.CurrentHP = m.MaxHP()
	m.ClearStatusEffects()
}
