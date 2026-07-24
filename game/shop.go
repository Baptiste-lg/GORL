package game

import "math/rand"

// Shop is an NPC that sells items on each floor.
type Shop struct {
	X, Y     int
	Items    []*Item // nil entries mean "sold out"
	Prices   []int
	Selected int
}

// NewShop creates a shop with randomized inventory for the given floor.
func NewShop(x, y, floor int, rng *rand.Rand) *Shop {
	// Generate 4 items: weapon, armor, potion, scroll
	// Bias rarity slightly up for shops (+0.1 bonus)
	bonus := FloorRarityBonus(floor) + 0.1

	items := []*Item{
		generateShopWeapon(rng, bonus),
		generateShopArmor(rng, bonus),
		GeneratePotion(rng, floor),
		GenerateScroll(rng, floor),
	}

	prices := make([]int, len(items))
	for i, item := range items {
		prices[i] = itemPrice(item, floor)
	}

	return &Shop{
		X:      x,
		Y:      y,
		Items:  items,
		Prices: prices,
	}
}

// Buy attempts to purchase the selected item. Returns true on success.
func (s *Shop) Buy(idx int, player *Player) bool {
	if idx < 0 || idx >= len(s.Items) {
		return false
	}
	if s.Items[idx] == nil {
		return false // already sold
	}
	if player.Gold < s.Prices[idx] {
		return false // can't afford
	}
	if player.Inventory.IsFull() {
		return false // no room
	}

	player.Gold -= s.Prices[idx]
	player.Inventory.Add(s.Items[idx])
	s.Items[idx] = nil // mark as sold
	return true
}

// HasItems returns true if the shop has any unsold items.
func (s *Shop) HasItems() bool {
	for _, item := range s.Items {
		if item != nil {
			return true
		}
	}
	return false
}

func generateShopWeapon(rng *rand.Rand, bonus float64) *Item {
	rarity := RollRarity(rng, bonus)
	b := int(rarity) + 1 + rng.Intn(int(rarity)+1)
	affixes := RollAffixes(rng, ItemWeapon, rarity)
	name := genWeaponName(rng, rarity)
	if len(affixes) > 0 {
		name = AffixNames(affixes) + " " + name
	}
	return &Item{
		Name:     name,
		Type:     ItemWeapon,
		Rarity:   rarity,
		BonusSTR: b,
		BonusDEX: rng.Intn(int(rarity) + 1),
		Affixes:  affixes,
	}
}

func generateShopArmor(rng *rand.Rand, bonus float64) *Item {
	rarity := RollRarity(rng, bonus)
	b := int(rarity) + 1 + rng.Intn(int(rarity)+1)
	affixes := RollAffixes(rng, ItemArmor, rarity)
	name := genArmorName(rng, rarity)
	if len(affixes) > 0 {
		name = AffixNames(affixes) + " " + name
	}
	return &Item{
		Name:     name,
		Type:     ItemArmor,
		Rarity:   rarity,
		BonusVIT: b,
		BonusDEX: rng.Intn(int(rarity) + 1),
		Affixes:  affixes,
	}
}
