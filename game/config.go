package game

import (
	"tbrpg/models"
)

type GameSettings struct {
	XPPerMonsterLevel           int          `json:"xp_per_monster_level"`
	XPToLevelUp                 []int        `json:"xp_to_level_up"`
	MoveLevelBonusPct           int          `json:"move_level_bonus_percent"`
	MaxRoomsPerLevel            int          `json:"max_rooms_per_level"`
	MaxEquippedMoves            int          `json:"max_equipped_moves"`
	MaxMoveLevel                int          `json:"max_move_level"`
	ManaRegenPerTurn            int          `json:"mana_regen_per_turn"`
	ManaRestoreBetweenFightsPct float32      `json:"mana_restore_between_fights_pct"`
}

type GameConfig struct {
	HeroTemplate     models.Hero
	MonsterTemplates []models.Monster
	Moves            map[string]models.MoveDefinition
	Settings         GameSettings
	EventTemplates   []models.Event
}
