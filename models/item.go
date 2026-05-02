package models

type ItemType string

const (
	Armor      ItemType = "armor"
	Weapon     ItemType = "weapon"
	Trinket    ItemType = "trinket"
	Consumable ItemType = "consumable"
)

// pretty self explanatory
// Since items should affects stats, there is a map of stat affected and how much they are affected
type Item struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Type        ItemType `json:"item_type"`
	DropRate    int      `json:"drop_rate"`
	Description string   `json:"description"`
	Price       int      `json:"price"`

	StatsAffected map[StatType]int `json:"stat_affected"`
}
