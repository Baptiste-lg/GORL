package game

import "math/rand"

// goldForKill calculates gold dropped by a killed enemy.
func goldForKill(baseXP, floor int, rng *rand.Rand) int {
	base := baseXP/2 + floor
	variance := rng.Intn(max(1, base/3))
	return base + variance
}

// itemPrice calculates the shop price for an item.
func itemPrice(item *Item, floor int) int {
	basePrice := 20 + int(item.Rarity)*30
	if item.Type == ItemWeapon || item.Type == ItemArmor {
		basePrice += 15
	}
	return basePrice + floor*5
}
