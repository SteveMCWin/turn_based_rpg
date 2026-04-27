package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
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
	ManualPointsOnLevelUp        int     `json:"manual_points_on_level_up"`
	RandomPointsOnLevelUp        int     `json:"random_points_on_level_up"`
}

type GameConfig struct {
	Moves            map[string]models.MoveDefinition
	Items            map[string]models.Item
	HeroTemplates    []models.Hero
	MonsterTemplates []models.Monster
	Settings         GameSettings
	EventTemplates   []models.Event
}

func LoadConfig(configDir string) (*GameConfig, error) {
	config := &GameConfig{}

	var moveList []models.MoveDefinition
	if err := loadJSON(configDir+"/moves.json", &moveList); err != nil {
		return nil, fmt.Errorf("moves config: %w", err)
	}
	config.Moves = make(map[string]models.MoveDefinition, len(moveList))
	for _, m := range moveList {
		config.Moves[m.ID] = m
	}

	var itemList []models.Item
	if err := loadJSON(configDir+"/items.json", &itemList); err != nil {
		return nil, fmt.Errorf("items config: %w", err)
	}
	config.Items = make(map[string]models.Item, len(itemList))
	for _, item := range itemList {
		config.Items[item.Id] = item
	}

	if err := loadJSON(configDir+"/hero.json", &config.HeroTemplates); err != nil {
		return nil, fmt.Errorf("hero config: %w", err)
	}
	for hero_idx := range config.HeroTemplates {
		for _, item_id := range config.HeroTemplates[hero_idx].ItemPool {
			config.HeroTemplates[hero_idx].EquipItem(config.Items[item_id])
		}
		// Clear item_pool so starting gear doesn't also appear in the player's inventory
		config.HeroTemplates[hero_idx].ItemPool = nil
	}

	if err := loadJSON(configDir+"/monsters.json", &config.MonsterTemplates); err != nil {
		return nil, fmt.Errorf("monsters config: %w", err)
	}
	for monster_idx := range config.MonsterTemplates {
		for _, item_id := range config.MonsterTemplates[monster_idx].ItemPool {
			item := config.Items[item_id]
			if rand.Intn(100) < item.DropRate {
				config.MonsterTemplates[monster_idx].EquipItem(item)
				break
			}
		}
	}


	if err := loadJSON(configDir+"/game.json", &config.Settings); err != nil {
		return nil, fmt.Errorf("game settings: %w", err)
	}

	// events.json is optional
	if err := loadJSON(configDir+"/events.json", &config.EventTemplates); err != nil {
		return nil, fmt.Errorf("events not loaded: %w", err)
	}

	return config, nil
}

func loadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
