package game

import (
	"encoding/json"
	"fmt"
	"os"

	"tbrpg/models"
)

type GameSettings struct {
	XPPerMonsterLevel            int     `json:"xp_per_monster_level"`
	XPToLevelUp                  []int   `json:"xp_to_level_up"`
	MoveLevelBonusPct            int     `json:"move_level_bonus_percent"`
	MaxRoomsPerLevel             int     `json:"max_rooms_per_level"`
	MaxEquippedMoves             int     `json:"max_equipped_moves"`
	MaxMoveLevel                 int     `json:"max_move_level"`
	ManaRegenPerTurn             int     `json:"mana_regen_per_turn"`
	ManaRestoreBetweenFightsPct  float32 `json:"mana_restore_between_fights_pct"`
	PercentChanceMonsterLevelsUp int     `json:"pct_chance_monster_lvl_up"`
}

type GameConfig struct {
	HeroTemplate     models.Hero
	MonsterTemplates []models.Monster
	Moves            map[string]models.MoveDefinition
	Settings         GameSettings
	EventTemplates   []models.Event
}

func LoadConfig(configDir string) (*GameConfig, error) {
	cfg := &GameConfig{}

	if err := loadJSON(configDir+"/hero.json", &cfg.HeroTemplate); err != nil {
		return nil, fmt.Errorf("hero config: %w", err)
	}

	if err := loadJSON(configDir+"/monsters.json", &cfg.MonsterTemplates); err != nil {
		return nil, fmt.Errorf("monsters config: %w", err)
	}

	var moveList []models.MoveDefinition
	if err := loadJSON(configDir+"/moves.json", &moveList); err != nil {
		return nil, fmt.Errorf("moves config: %w", err)
	}
	cfg.Moves = make(map[string]models.MoveDefinition, len(moveList))
	for _, m := range moveList {
		cfg.Moves[m.ID] = m
	}

	if err := loadJSON(configDir+"/game.json", &cfg.Settings); err != nil {
		return nil, fmt.Errorf("game settings: %w", err)
	}

	// events.json is optional
	if err := loadJSON(configDir+"/events.json", &cfg.EventTemplates); err != nil {
		fmt.Printf("warning: events config not loaded: %v\n", err)
	}

	return cfg, nil
}

func loadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
