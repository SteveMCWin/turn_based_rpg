package game

import (
	"fmt"
	"math/rand"
	"slices"

	"tbrpg/models"
)

// Response for the client after it requests to update the game state in battle
// e.g. when a player makes a move, this is the response from the server
type BattleResult struct {
	GameState    *Game               `json:"game_state,omitempty"`
	BattleOver   bool                `json:"battle_over"`
	PlayerWon    bool                `json:"player_won"`
	WasBoss      bool                `json:"was_boss,omitempty"`
	MonsterName  string              `json:"monster_name,omitempty"`
	FloorReached int                 `json:"floor_reached,omitempty"`
	LearnedMove  *models.LearnedMove `json:"learned_move,omitempty"`
	ItemAcquired *models.Item        `json:"item_acquired,omitempty"`
	NewLogLines  []string            `json:"new_log_lines,omitempty"`
}

func (g *Game) SubmitPlayerMove(moveID string) (*BattleResult, error) {
	if !g.IsInBattle {
		return nil, fmt.Errorf("not in battle")
	}
	if g.WaitingForMonster {
		return nil, fmt.Errorf("waiting for monster turn")
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

	// scale the move strenght by it's level and scaling amount
	scaled_value := int(float64(move_def.BaseValue) * (1.0 + float64(moveLevel-1)*float64(g.Settings.MoveLevelBonusPct)/100.0))
	effMove := move_def
	effMove.BaseValue = scaled_value

	var logLines []string

	preHeroHP := hero.CurrentHP

	// apply move and log
	hero.CurrentMana -= move_def.CostAmount
	applyMove(effMove, &hero.Entity, &monster.Entity)
	logLines = append(logLines, fmt.Sprintf("You use %s. %s", move_def.Name, describeMoveResult(effMove, &hero.Entity, &monster.Entity, preHeroHP)))

	// check if fight is over
	if !monster.IsAlive() {
		monster.CurrentHP = 0
		result := g.endBattle(true)
		result.NewLogLines = logLines
		return result, nil
	}

	hero.CurrentMana = min(hero.CurrentMana+g.Settings.ManaRegenPerTurn, hero.MaxMana())
	hero.TickStatusEffects()

	// switch to monsters turn
	g.WaitingForMonster = true
	return &BattleResult{GameState: g, NewLogLines: logLines}, nil
}

// handle the monsters turn in battle
// called by the client automatically after the player makes a move
func (g *Game) SubmitMonsterMove() (*BattleResult, error) {
	if !g.IsInBattle {
		return nil, fmt.Errorf("not in battle")
	}

	if !g.WaitingForMonster {
		return nil, fmt.Errorf("not waiting for monster turn")
	}

	room := g.CurrentRoom()
	monster := room.Encounter.Monster
	hero := &g.Player

	var logLines []string

	monster.TickStatusEffects()

	if !monster.IsAlive() {
		monster.CurrentHP = 0
		g.WaitingForMonster = false
		result := g.endBattle(true)
		result.NewLogLines = logLines
		return result, nil
	}

	monster.CurrentMana = min(monster.CurrentMana+g.Settings.ManaRegenPerTurn, monster.MaxMana())

	monsterMoveID, err := g.pickMonsterMove()
	if err != nil {
		return nil, err
	}

	monsterMoveDef := g.AllMoves[monsterMoveID]
	monster.CurrentMana -= monsterMoveDef.CostAmount
	preMonsterHP := monster.CurrentHP
	applyMove(monsterMoveDef, &monster.Entity, &hero.Entity)
	logLines = append(logLines, fmt.Sprintf("%s uses %s. %s", monster.Name, monsterMoveDef.Name, describeMoveResult(monsterMoveDef, &monster.Entity, &hero.Entity, preMonsterHP)))

	g.WaitingForMonster = false

	if !hero.IsAlive() {
		hero.CurrentHP = 0
		result := g.endBattle(false)
		result.NewLogLines = logLines
		return result, nil
	}

	return &BattleResult{GameState: g, NewLogLines: logLines}, nil
}

// Takes into account monsters stats and heroes stats
// makes decision based on that and what the move does
func (g *Game) pickMonsterMove() (string, error) {
	monster := g.CurrentRoom().Encounter.Monster

	if len(monster.Moves) == 0 {
		return "", fmt.Errorf("monster %q has no moves", monster.Name)
	}

	// how much hp does the monster have on a scale from 0.0 to 1.0 (1.0 is max hp)
	monsterHpPct := float64(monster.CurrentHP) / float64(max(1, monster.MaxHP()))
	hero := &g.Player

	hasActiveBuff := false
	for _, se := range monster.StatusEffects {
		if se.Type == models.StatModifier && se.BaseDelta > 0 && !se.IsEnvironmental && se.TurnsToActivate <= 0 {
			hasActiveBuff = true
			break
		}
	}

	heroHasDebuff := false
	for _, se := range hero.StatusEffects {
		if se.Type == models.StatModifier && se.BaseDelta < 0 && !se.IsEnvironmental && se.TurnsToActivate <= 0 {
			heroHasDebuff = true
			break
		}
	}

	type candidate struct {
		id     string
		weight int
	}
	var candidates []candidate

	for _, m := range monster.Moves {
		def, ok := g.AllMoves[m.MoveID]
		if !ok {
			continue
		}

		if def.CostAmount > monster.CurrentMana {
			continue
		}

		weight := 1
		switch def.Intent {
		case models.IntentDamage:
			// dealing damage is always decent
			weight = 4

		case models.IntentHeal:
			// strong chance of healing if low on hp
			if monsterHpPct < 0.35 {
				weight = 8
			} else if monsterHpPct < 0.5 {
				weight = 5
			} else {
				weight = 0
			}

		case models.IntentBuff:
			// buff is good if not applied alreaady
			if !hasActiveBuff {
				weight = 4
			} else {
				weight = 0
			}
		case models.IntentDebuff:
			// prioritize damage or heal if low on hp or if player is already debuffed
			if monsterHpPct < 0.35 || heroHasDebuff {
				weight = 0
			} else {
				weight = 3
			}
		}

		if weight > 0 {
			candidates = append(candidates, candidate{m.MoveID, weight})
		}
	}

	// fallback
	if len(candidates) == 0 {
		return monster.Moves[rand.Intn(len(monster.Moves))].MoveID, nil
	}

	// if a damage move will kill the hero, ignore previous weights
	for _, c := range candidates {
		def := g.AllMoves[c.id]
		if def.Intent == models.IntentDamage && calcDamage(def, &monster.Entity, &hero.Entity) >= hero.CurrentHP {
			return c.id, nil
		}
	}

	total := 0
	for _, c := range candidates {
		total += c.weight
	}

	// candidates with higher weight have a better chance of bringing the rand num below 0
	r := rand.Intn(total)
	for _, c := range candidates {
		r -= c.weight
		if r < 0 {
			return c.id, nil
		}
	}

	// another fallback
	return candidates[len(candidates)-1].id, nil
}

// return stat value based on stat type
func scalingStatValue(stats models.Stats, stat models.StatType) int {
	switch stat {
	case models.AttackStat:
		return stats.Attack
	case models.MagicStat:
		return stats.Magic
	case models.DefenseStat:
		return stats.Defense
	case models.HealthStat:
		return stats.Health
	case models.ManaStat:
		return stats.Mana
	}
	return 0
}

// returns amount of hp the defender will lose based on the move selected and attackers and defenders stats
func calcDamage(move models.MoveDefinition, attacker, defender *models.Entity) int {
	eff := attacker.EffectiveStats()
	defEff := defender.EffectiveStats()
	power := int(float32(scalingStatValue(eff, move.ScalingStat)) * move.ScalingStatFactor * float32(move.BaseValue) / 100)
	if move.MoveType == models.Physical {
		return max(1, power-defEff.Defense)
	}
	return max(1, power)
}

// returns amount of hp the caster will gain from a healing move
func calcHeal(move models.MoveDefinition, caster *models.Entity) int {
	eff := caster.EffectiveStats()
	return int(float32(scalingStatValue(eff, move.ScalingStat)) * move.ScalingStatFactor * float32(move.BaseValue) / 100)
}

// calc damage/heal and effect of a move based on move selected, attacker and defender
func applyMove(move models.MoveDefinition, attacker, defender *models.Entity) {
	switch move.Intent {
	case models.IntentDamage:
		dmg := calcDamage(move, attacker, defender)
		defender.CurrentHP = max(0, defender.CurrentHP-dmg)

	case models.IntentHeal:
		heal := calcHeal(move, attacker)
		attacker.CurrentHP = min(attacker.CurrentHP+heal, attacker.MaxHP())
	}

	for i := range move.Effects {
		target := defender
		if move.Effects[i].Target == models.TargetSelf {
			target = attacker
		}
		delta := computeEffectDelta(move.Effects[i], attacker, move.ScalingStat)
		addEffect(move.Effects[i], target, delta)
	}
}

// compute how strong the effect of a move is based on attacker stats and stat scaling
// if no scaling factor, use base delta
func computeEffectDelta(effect *models.Effect, attacker *models.Entity, stat models.StatType) int {
	if effect.ScaleFactor != 0 {
		eff := attacker.EffectiveStats()
		return int(float32(scalingStatValue(eff, stat)) * effect.ScaleFactor)
	}

	return effect.BaseDelta
}

func addEffect(effect *models.Effect, target *models.Entity, delta int) {
	se := models.StatusEffect{
		Effect:          *effect,
		TurnsRemaining:  effect.Duration,
		TurnsToActivate: effect.ActivationDelay,
	}
	se.BaseDelta = delta

	target.StatusEffects = append(target.StatusEffects, se)
}

func addEnvironmentEffect(effect *models.Effect, target *models.Entity) {
	target.StatusEffects = append(target.StatusEffects, models.StatusEffect{
		Effect:          *effect,
		TurnsRemaining:  effect.Duration,
		TurnsToActivate: effect.ActivationDelay,
		IsEnvironmental: true,
	})
}

// logging
func describeMoveResult(move models.MoveDefinition, attacker, defender *models.Entity, preAttackerHP int) string {
	switch move.Intent {
	case models.IntentDamage:
		return fmt.Sprintf("Deals %d damage.", calcDamage(move, attacker, defender))
	case models.IntentHeal:
		return fmt.Sprintf("Restores %d HP.", attacker.CurrentHP-preAttackerHP)
	}

	for i := range move.Effects {
		e := move.Effects[i]
		switch e.Type {
		case models.StatModifier:
			sign := ""
			if e.BaseDelta > 0 {
				sign = "+"
			}
			return fmt.Sprintf("%s %s%d for %d turns.", e.StatAffected, sign, e.BaseDelta, e.Duration)
		case models.DamageOverTime:
			if e.BaseDelta > 0 {
				return fmt.Sprintf("Deals %d damage.", e.BaseDelta)
			}
			return fmt.Sprintf("Restores %d HP.", -e.BaseDelta)
		}
	}
	return ""
}

// handles end of battle and returns updated game state
func (g *Game) endBattle(playerWon bool) *BattleResult {
	result := &BattleResult{BattleOver: true, PlayerWon: playerWon}
	room := g.CurrentRoom()
	g.IsInBattle = false

	var logLines []string

	if playerWon {
		monster := room.Encounter.Monster
		logLines = append(logLines, fmt.Sprintf("You defeated %s!", monster.Name))

		xpGain := g.Settings.XPPerMonsterLevel * monster.Level
		levelsGained := g.Player.AddXP(xpGain, g.Settings.XPToLevelUp)
		logLines = append(logLines, fmt.Sprintf("Gained %d XP.", xpGain))

		if levelsGained > 0 {
			logLines = append(logLines, fmt.Sprintf("Level up! Now Lv.%d", g.Player.Level))
			g.PendingLevelUp = &PendingAllocation{
				ManualPoints: g.Settings.ManualPointsOnLevelUp * levelsGained,
				RandomPoints: g.Settings.RandomPointsOnLevelUp * levelsGained,
			}
		}

		// handle learning a move after defating the monster
		learned := g.learnFromMonster()
		if learned != nil {
			moveName := g.AllMoves[learned.MoveID].Name
			logLines = append(logLines, fmt.Sprintf("Learned: %s (Lv.%d)", moveName, learned.Level))
			result.LearnedMove = learned
		}

		// handle getting item from monster
		item_acquired := g.lootMonster()
		if item_acquired != nil {
			logLines = append(logLines, fmt.Sprintf("Got an item: %s", item_acquired.Name))
			result.ItemAcquired = item_acquired
		}

		// restore part of max mana after completing battle
		if g.Player.MaxMana() > 0 {
			restore := int(float32(g.Player.MaxMana()) * g.Settings.ManaRestoreBetweenFightsPct)
			g.Player.CurrentMana = min(g.Player.CurrentMana+restore, g.Player.MaxMana())
		}

		g.Player.ClearStatusEffects()

		// 🤑🤑🤑
		// (get gold in the range from min to max defined in settings)
		max_g := g.Settings.MaxGoldAfterBattle
		min_g := g.Settings.MinGoldAfterBattle
		gold_looted := rand.Intn(max_g-min_g) + min_g
		// get double the gold if monster was a boss
		isBoss := room.Encounter.Kind == models.EncounterKindBoss
		if isBoss {
			gold_looted *= 2
			result.WasBoss = true
		}
		g.Player.CurrentGold += gold_looted

		g.LastBattleResult = &BattleResult{
			PlayerWon:   true,
			WasBoss:     isBoss,
			MonsterName: monster.Name,
			LearnedMove: learned,
		}

		g.CompleteRoom(g.CurrentRoomID)
		g.CurrentRoomID = ""

	} else {
		logLines = append(logLines, "You were defeated...")
		g.Player.ClearStatusEffects()
		g.Player.CurrentHP = g.Player.MaxHP()
		g.Player.CurrentMana = g.Player.MaxMana()

		floorReached := 0
		if g.IsEndless {
			if fi, _, err := models.GetFloorRoomIdx(room.Id); err == nil {
				floorReached = fi + 1
			}
		}
		g.LastBattleResult = &BattleResult{
			PlayerWon:    false,
			MonsterName:  room.Encounter.Monster.Name,
			FloorReached: floorReached,
		}

		g.CurrentRoomID = ""
	}

	result.NewLogLines = logLines
	result.GameState = g
	return result
}
