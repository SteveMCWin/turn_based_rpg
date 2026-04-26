package models

// The entity field is the base entity struct with stats and status effects
// learned moves are ids of all the moves a player can equip at the moment
// equipped moves is self explanatory :^)
type Hero struct {
	Entity
	Name          string        `json:"name"`
	LearnedMoves  []LearnedMove `json:"learned_moves"`
	EquippedMoves []string      `json:"equipped_moves"`
	MaxMoveLevel  int           `json:"max_move_level"`
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

func (h *Hero) LearnMove(moveID string, level int) LearnedMove {
	for i := range h.LearnedMoves {
		if h.LearnedMoves[i].MoveID == moveID {
			if h.LearnedMoves[i].Level < h.MaxMoveLevel {
				h.LearnedMoves[i].Level++
			}
			return h.LearnedMoves[i]
		}
	}
	lm := LearnedMove{MoveID: moveID, Level: min(level, h.MaxMoveLevel)}
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

// Monster represents an enemy in the gauntlet.
type Monster struct {
	Entity
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Moves      []LearnedMove `json:"learned_moves"`
	IsDefeated bool          `json:"is_defeated"`
}

func (m *Monster) Init() {
	m.CurrentHP = m.Stats.Health
	m.CurrentMana = m.Stats.Mana
	m.StatusEffects = make([]StatusEffect, 0)
}

// func (m *Monster) LevelUp() {
// 	m.Level++
// 	m.LevelUpStats(m.Level)
// }

func (m *Monster) ResetForBattle() {
	m.CurrentHP = m.MaxHP()
	m.ClearStatusEffects()
}
