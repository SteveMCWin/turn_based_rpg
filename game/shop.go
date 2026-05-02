package game

import (
	"fmt"
)

func (g *Game) BuyItem(itemId string) error {
	if g.IsInBattle {
		return fmt.Errorf("cannot shop during battle")
	}
	idx := -1
	for i, item := range g.Shop.Items {
		if item.Id == itemId {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("item not available in shop")
	}
	item := g.Shop.Items[idx]
	if g.Player.CurrentGold < item.Price {
		return fmt.Errorf("not enough gold")
	}
	g.Player.CurrentGold -= item.Price
	g.Player.ItemPool = append(g.Player.ItemPool, item.Id)
	g.Shop.Items = append(g.Shop.Items[:idx], g.Shop.Items[idx+1:]...)
	return nil
}

func (g *Game) SellItem(itemId string) error {
	if g.IsInBattle {
		return fmt.Errorf("cannot shop during battle")
	}
	found := false
	for i, id := range g.Player.ItemPool {
		if id == itemId {
			g.Player.ItemPool = append(g.Player.ItemPool[:i], g.Player.ItemPool[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("item not in inventory")
	}
	item, ok := g.Config.Items[itemId]
	if !ok {
		return fmt.Errorf("unknown item")
	}
	g.Player.CurrentGold += int(float32(item.Price) * g.Settings.SellModifier)
	return nil
}
