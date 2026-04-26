package game

import (
	"fmt"
	"math/rand"
	"slices"

	"tbrpg/models"
)

type BattleResult struct {
	GameState    *Game               `json:"game_state,omitempty"`
	BattleOver   bool                `json:"battle_over"`
	PlayerWon    bool                `json:"player_won"`
	MonsterName  string              `json:"monster_name,omitempty"`
	LearnedMove  *models.LearnedMove `json:"learned_move,omitempty"`
	ItemAcquired *models.Item         `json:"item_acquired,omitempty"`
}

func (g *Game) SubmitPlayerMove(moveID string) (*BattleResult, error) {
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

	hero.CurrentMana -= move_def.CostAmount

	scaled_value := int(float64(move_def.BaseValue) * (1.0 + float64(moveLevel-1)*float64(g.Settings.MoveLevelBonusPct)/100.0))
	effMove := move_def
	effMove.BaseValue = scaled_value

	// =========================
	// ======= Hero turn =======
	// =========================

	hero.TickStatusEffects()
	applyMove(effMove, &hero.Entity, &monster.Entity)
	new_log_line := fmt.Sprintf("You use %s. %s", move_def.Name, describeMoveResult(effMove, &hero.Entity, &monster.Entity))
	g.BattleLog = append(g.BattleLog, new_log_line)

	hero.CurrentMana = min(hero.CurrentMana+g.Settings.ManaRegenPerTurn, hero.MaxMana())

	if !monster.IsAlive() {
		monster.CurrentHP = 0
		return g.endBattle(true), nil
	}

	// ============================
	// ======= Monster turn =======
	// ============================

	monster.TickStatusEffects()

	if !monster.IsAlive() {
		monster.CurrentHP = 0
		return g.endBattle(true), nil
	}

	monster.CurrentMana = min(monster.CurrentMana+g.Settings.ManaRegenPerTurn, monster.MaxMana())

	monsterMoveID := g.pickMonsterMove()
	monsterMoveDef, ok := g.AllMoves[monsterMoveID]
	if !ok && len(monster.Moves) > 0 {
		monsterMoveDef = g.AllMoves[monster.Moves[0].MoveID]
	}
	applyMove(monsterMoveDef, &monster.Entity, &hero.Entity)
	g.BattleLog = append(g.BattleLog, fmt.Sprintf("%s uses %s. %s", monster.Name, monsterMoveDef.Name, describeMoveResult(monsterMoveDef, &monster.Entity, &hero.Entity)))

	if !hero.IsAlive() {
		hero.CurrentHP = 0
		return g.endBattle(false), nil
	}

	return &BattleResult{GameState: g}, nil
}

func (g *Game) pickMonsterMove() string {
	monster := g.CurrentRoom().Encounter.Monster
	pool := monster.Moves
	monsterHP := float64(monster.CurrentHP) / float64(max(1, monster.MaxHP()))
	// heroHP := float64(hero.CurrentHP) / float64(max(1, hero.MaxHP()))

	type candidate struct {
		id     string
		weight int
	}
	var candidates []candidate

	for _, m := range pool {
		def, ok := g.AllMoves[m.MoveID]
		if !ok {
			continue
		}
		if def.CostAmount > monster.CurrentMana {
			continue
		}

		weight := 1
		switch {
		case def.Primary == models.PrimaryDamage:
			weight = 4
		case def.Primary == models.PrimaryHeal:
			if monsterHP < 0.35 {
				weight = 6
			} else {
				weight = 0
			}
			// TODO: check if hero has negative stat effects and if this move applies a debuff, give it a weight
		}

		if weight > 0 {
			candidates = append(candidates, candidate{m.MoveID, weight})
		}
	}

	if len(candidates) == 0 {
		return pool[rand.Intn(len(pool))].MoveID
	}

	total := 0
	for _, c := range candidates {
		total += c.weight
	}
	r := rand.Intn(total)
	for _, c := range candidates {
		r -= c.weight
		if r < 0 {
			return c.id
		}
	}
	return candidates[len(candidates)-1].id
}

func applyMove(move models.MoveDefinition, attacker, defender *models.Entity) {
	eff := attacker.EffectiveStats()
	defEff := defender.EffectiveStats()

	switch move.Primary {
	case models.PrimaryDamage:
		var dmg int
		if move.MoveType == models.Physical {
			dmg = max(1, eff.Attack*move.BaseValue/100-defEff.Defense)
		} else {
			dmg = max(1, eff.Magic*move.BaseValue/100)
		}
		defender.CurrentHP = max(0, defender.CurrentHP-dmg)
	case models.PrimaryHeal:
		heal := eff.Magic * move.BaseValue / 100
		attacker.CurrentHP = min(attacker.CurrentHP+heal, attacker.MaxHP())
	}

	if len(move.Effects) > 0 {
		for i := range move.Effects {
			target := defender
			if move.Effects[i].Target == models.TargetSelf {
				target = attacker
			}
			applyEffect(move.Effects[i], target)
		}
	}
}

func applyEffect(effect *models.Effect, target *models.Entity) {
	target.StatusEffects = append(target.StatusEffects, models.StatusEffect{
		Effect:          *effect,
		TurnsRemaining:  effect.Duration,
		TurnsToActivate: effect.ActivationDelay,
	})
}

func describeMoveResult(move models.MoveDefinition, attacker, defender *models.Entity) string {
	eff := attacker.EffectiveStats()
	defEff := defender.EffectiveStats()

	switch move.Primary {
	case models.PrimaryDamage:
		var dmg int
		if move.MoveType == models.Physical {
			dmg = max(1, eff.Attack*move.BaseValue/100-defEff.Defense)
		} else {
			dmg = max(1, eff.Magic*move.BaseValue/100)
		}
		return fmt.Sprintf("Deals %d damage.", dmg)
	case models.PrimaryHeal:
		return fmt.Sprintf("Restores %d HP.", eff.Magic*move.BaseValue/100)
	}

	for i := range move.Effects {
		e := move.Effects[i]
		switch e.Type {
		case models.StatModifier:
			sign := ""
			if e.Delta > 0 {
				sign = "+"
			}
			return fmt.Sprintf("%s %s%d for %d turns.", e.StatAffected, sign, e.Delta, e.Duration)
		case models.DamageOverTime:
			if e.Delta > 0 {
				return fmt.Sprintf("Deals %d damage.", e.Delta)
			}
			return fmt.Sprintf("Restores %d HP.", -e.Delta)
		}
	}
	return ""
}

func (g *Game) endBattle(playerWon bool) *BattleResult {
	result := &BattleResult{BattleOver: true, PlayerWon: playerWon}
	room := g.CurrentRoom()
	g.IsInBattle = false

	if playerWon {
		monster := room.Encounter.Monster
		monster.IsDefeated = true
		g.BattleLog = append(g.BattleLog, fmt.Sprintf("You defeated %s!", monster.Name))

		xpGain := g.Settings.XPPerMonsterLevel * monster.Level
		levelsGained := g.Player.AddXP(xpGain, g.Settings.XPToLevelUp)
		g.BattleLog = append(g.BattleLog, fmt.Sprintf("Gained %d XP.", xpGain))

		if levelsGained > 0 {
			g.BattleLog = append(g.BattleLog, fmt.Sprintf("Level up! Now Lv.%d", g.Player.Level))
			g.PendingLevelUp = &PendingAllocation{
				ManualPoints: g.Settings.ManualPointsOnLevelUp * levelsGained,
				RandomPoints: g.Settings.RandomPointsOnLevelUp * levelsGained,
			}
		}

		learned := g.learnRandomMove()
		if learned != nil {
			moveName := g.AllMoves[learned.MoveID].Name
			g.BattleLog = append(g.BattleLog, fmt.Sprintf("Learned: %s (Lv.%d)", moveName, learned.Level))
			result.LearnedMove = learned
		}

		item_acquired := g.getRandomItem()
		if item_acquired != nil {
			item_name := item_acquired.Name
			g.BattleLog = append(g.BattleLog, fmt.Sprintf("Got an item: %s", item_name))
			result.ItemAcquired = item_acquired
		}

		// Restore a portion of hero mana between fights
		if g.Player.MaxMana() > 0 {
			restore := int(float32(g.Player.MaxMana()) * g.Settings.ManaRestoreBetweenFightsPct)
			g.Player.CurrentMana = min(g.Player.CurrentMana+restore, g.Player.MaxMana())
		}
		g.Player.ClearStatusEffects()

		g.LastBattleResult = &BattleResult{
			PlayerWon:   true,
			MonsterName: monster.Name,
			LearnedMove: learned,
		}

		g.CompleteRoom(g.CurrentRoomID)
		g.CurrentRoomID = ""
	} else {
		g.BattleLog = append(g.BattleLog, "You were defeated...")
		if room != nil && room.Encounter.Monster != nil {
			room.Encounter.Monster.ResetForBattle()
		}
		g.Player.ClearStatusEffects()

		g.LastBattleResult = &BattleResult{
			PlayerWon:   false,
			MonsterName: room.Encounter.Monster.Name,
		}

		g.CurrentRoomID = ""
	}

	result.GameState = g
	return result
}
