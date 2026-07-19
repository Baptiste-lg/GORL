package game

import "math/rand"

// ItemType classifies items.
type ItemType int

const (
	ItemWeapon ItemType = iota
	ItemArmor
	ItemPotion
	ItemScroll
)

// Rarity determines item power and color.
type Rarity int

const (
	RarityCommon   Rarity = iota // white
	RarityUncommon               // green
	RarityRare                   // blue
	RarityEpic                   // purple
)

func (r Rarity) Color() string {
	switch r {
	case RarityUncommon:
		return "#44ff44"
	case RarityRare:
		return "#4488ff"
	case RarityEpic:
		return "#cc44ff"
	default:
		return "#cccccc"
	}
}

func (r Rarity) Name() string {
	switch r {
	case RarityUncommon:
		return "Uncommon"
	case RarityRare:
		return "Rare"
	case RarityEpic:
		return "Epic"
	default:
		return "Common"
	}
}

// ScrollKind identifies scroll subtypes.
type ScrollKind int

const (
	ScrollSpeed ScrollKind = iota
	ScrollShield
)

// Item is something the player can pick up, equip, or use.
type Item struct {
	Name       string
	Type       ItemType
	Rarity     Rarity
	BonusSTR   int
	BonusDEX   int
	BonusVIT   int
	BonusLCK   int
	ScrollKind ScrollKind
	// World position (0,0 if in inventory)
	X, Y int
}

// Glyph returns the ASCII character for items on the ground.
func (it *Item) Glyph() string {
	switch it.Type {
	case ItemWeapon:
		return "/"
	case ItemArmor:
		return "["
	case ItemPotion:
		return "!"
	case ItemScroll:
		return "?"
	}
	return "*"
}

// --- Name generation ---

var weaponPrefixes = []string{"Rusty", "Sharp", "Keen", "Brutal", "Swift", "Heavy"}
var weaponBases = []string{"Sword", "Axe", "Dagger", "Mace", "Spear"}
var armorPrefixes = []string{"Worn", "Sturdy", "Hardened", "Enchanted", "Plated"}
var armorBases = []string{"Robe", "Leather", "Chainmail", "Plate", "Shield"}

func genWeaponName(rng *rand.Rand, rarity Rarity) string {
	prefix := weaponPrefixes[rng.Intn(len(weaponPrefixes))]
	base := weaponBases[rng.Intn(len(weaponBases))]
	if rarity >= RarityRare {
		prefix = "Enchanted"
	}
	return prefix + " " + base
}

func genArmorName(rng *rand.Rand, rarity Rarity) string {
	prefix := armorPrefixes[rng.Intn(len(armorPrefixes))]
	base := armorBases[rng.Intn(len(armorBases))]
	if rarity >= RarityRare {
		prefix = "Enchanted"
	}
	return prefix + " " + base
}

// --- Item generation ---

// RollRarity picks a rarity based on floor depth bonus.
func RollRarity(rng *rand.Rand, floorBonus float64) Rarity {
	roll := rng.Float64() + floorBonus
	switch {
	case roll > 0.95:
		return RarityEpic
	case roll > 0.80:
		return RarityRare
	case roll > 0.55:
		return RarityUncommon
	default:
		return RarityCommon
	}
}

// FloorRarityBonus returns the loot rarity bonus for a given floor.
func FloorRarityBonus(floor int) float64 {
	switch {
	case floor >= 15:
		return 0.40
	case floor >= 10:
		return 0.25
	case floor >= 5:
		return 0.10
	default:
		return 0.0
	}
}

// GenerateWeapon creates a random weapon.
func GenerateWeapon(rng *rand.Rand, floor int) *Item {
	rarity := RollRarity(rng, FloorRarityBonus(floor))
	bonus := int(rarity) + 1 + rng.Intn(int(rarity)+1)
	return &Item{
		Name:     genWeaponName(rng, rarity),
		Type:     ItemWeapon,
		Rarity:   rarity,
		BonusSTR: bonus,
		BonusDEX: rng.Intn(int(rarity) + 1),
	}
}

// GenerateArmor creates a random armor piece.
func GenerateArmor(rng *rand.Rand, floor int) *Item {
	rarity := RollRarity(rng, FloorRarityBonus(floor))
	bonus := int(rarity) + 1 + rng.Intn(int(rarity)+1)
	return &Item{
		Name:     genArmorName(rng, rarity),
		Type:     ItemArmor,
		Rarity:   rarity,
		BonusVIT: bonus,
		BonusDEX: rng.Intn(int(rarity) + 1),
	}
}

// GeneratePotion creates a health potion.
func GeneratePotion(rng *rand.Rand, floor int) *Item {
	rarity := RollRarity(rng, FloorRarityBonus(floor))
	name := "Health Potion"
	switch rarity {
	case RarityUncommon:
		name = "Greater Health Potion"
	case RarityRare:
		name = "Superior Health Potion"
	case RarityEpic:
		name = "Supreme Health Potion"
	}
	return &Item{
		Name:   name,
		Type:   ItemPotion,
		Rarity: rarity,
	}
}

// GenerateScroll creates a random scroll.
func GenerateScroll(rng *rand.Rand, floor int) *Item {
	rarity := RollRarity(rng, FloorRarityBonus(floor))
	kind := ScrollKind(rng.Intn(2))
	name := "Speed Scroll"
	if kind == ScrollShield {
		name = "Shield Scroll"
	}
	return &Item{
		Name:       rarity.Name() + " " + name,
		Type:       ItemScroll,
		Rarity:     rarity,
		ScrollKind: kind,
	}
}

// GenerateLoot creates a random item appropriate for the floor.
func GenerateLoot(rng *rand.Rand, floor int) *Item {
	roll := rng.Intn(100)
	switch {
	case roll < 30:
		return GenerateWeapon(rng, floor)
	case roll < 55:
		return GenerateArmor(rng, floor)
	case roll < 80:
		return GeneratePotion(rng, floor)
	default:
		return GenerateScroll(rng, floor)
	}
}

// PotionHealPercent returns how much % of max HP a potion heals.
func PotionHealPercent(rarity Rarity) float64 {
	switch rarity {
	case RarityUncommon:
		return 0.50
	case RarityRare:
		return 0.80
	case RarityEpic:
		return 1.00
	default:
		return 0.30
	}
}
