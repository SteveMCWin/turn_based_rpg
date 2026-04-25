package game

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"tbrpg/models"
)

type BattleResult struct {
	GameState   *Game               `json:"game_state"`
	BattleOver  bool                `json:"battle_over"`
	PlayerWon   bool                `json:"player_won"`
	LearnedMove *models.LearnedMove `json:"learned_move,omitempty"`
}

func (g *Game) SubmitPlayerMove(moveID string, cfg *GameConfig) (*BattleResult, error) {
	if !g.IsInBattle {
		return nil, fmt.Errorf("not in battle")
	}

	move_def, ok := g.AllMoves[moveID]
	if !ok {
		return nil, fmt.Errorf("move with id %s doesn't exist?", moveID)
	}

	room := g.CurrentRoom()

	monster := room.Encounter.Monster
	hero := &g.Player

	if !slices.Contains(hero.EquippedMoves, moveID) {
		return nil, fmt.Errorf("move %s is not equipped", moveID)
	}

	if move_def.CostAmount > hero.CurrentMana {
		return nil, fmt.Errorf("not enough mana")
	}

	moveLevel := hero.GetMoveLevel(moveID)
}
