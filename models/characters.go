package models

// The entity field is the base entity struct with stats and status effects
// learned moves are ids of all the moves a player can equip at the moment
// equipped moves and current gold are self explanatory :^)
type Hero struct {
	Entity
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	LearnedMoves  []LearnedMove `json:"learned_moves"`
	EquippedMoves []string      `json:"equipped_moves"`
	CurrentGold   int           `json:"current_gold"`
}

// Called at the start of the game
func (h *Hero) Reset() {
	h.CurrentHP = h.MaxHP()
	h.CurrentMana = h.MaxMana()
	h.StatusEffects = make([]StatusEffect, 0)
	if len(h.LearnedMoves) > 0 {
		h.EquippedMoves = make([]string, len(h.LearnedMoves))
		for i, m := range h.LearnedMoves {
			h.EquippedMoves[i] = m.MoveID
		}
	}
}

// Adds to the pool of learned moves. If already learned, level it up
func (h *Hero) LearnMove(moveID string, level int) LearnedMove {
	for i := range h.LearnedMoves {
		if h.LearnedMoves[i].MoveID == moveID {
			h.LearnedMoves[i].Level++
			return h.LearnedMoves[i]
		}
	}
	lm := LearnedMove{MoveID: moveID, Level: level}
	h.LearnedMoves = append(h.LearnedMoves, lm)
	return lm
}

func (h *Hero) GetMoveLevel(moveID string) int {
	for _, lm := range h.LearnedMoves {
		if lm.MoveID == moveID {
			return lm.Level
		}
	}
	return 0
}

type Monster struct {
	Entity
	ID   string        `json:"id"`
	Name string        `json:"name"`
	Moves []LearnedMove `json:"learned_moves"`
}

// called upon initialization and before every battle
func (m *Monster) Reset() {
	m.CurrentHP = m.MaxHP()
	m.CurrentMana = m.MaxMana()
	m.ClearStatusEffects()
}
