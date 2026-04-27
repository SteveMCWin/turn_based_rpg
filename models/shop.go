package models

import mrand "math/rand"

type Shop struct {
	Items []Item `json:"items"`
}

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

	mrand.Shuffle(len(equippables), func(i, j int) {
		equippables[i], equippables[j] = equippables[j], equippables[i]
	})
	if len(equippables) > 4 {
		equippables = equippables[:4]
	}

	return &Shop{Items: append(equippables, consumables...)}
}
