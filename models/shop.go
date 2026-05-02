package models

import "math/rand"

const NUM_RANDOM_ITEMS = 4

// Just a collection of items the shop has to offer
type Shop struct {
	Items []Item `json:"items"`
}

// The shop will contain 4 random equippable items and one of each consumable item
func NewShop(allItems map[string]Item) *Shop {
	var equippables []Item
	var consumables []Item

	for _, item := range allItems {
		switch item.Type {
		case Consumable:
			consumables = append(consumables, item)
		default:
			equippables = append(equippables, item)
		}
	}

	rand.Shuffle(len(equippables), func(i, j int) {
		equippables[i], equippables[j] = equippables[j], equippables[i]
	})
	if len(equippables) > NUM_RANDOM_ITEMS {
		equippables = equippables[:NUM_RANDOM_ITEMS]
	}

	return &Shop{Items: append(equippables, consumables...)}
}
