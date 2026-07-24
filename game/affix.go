package game

import "math/rand"

// AffixID identifies a specific affix type.
type AffixID int

const (
	// Weapon affixes
	AffixVampiric    AffixID = iota // Heal 1 HP on kill
	AffixBurning                    // +3 fire damage per hit
	AffixFreezing                   // 15% chance to skip enemy's next turn
	AffixVenomous                   // 20% chance to poison enemy
	AffixBerserker                  // +50% damage when below 30% HP
	AffixSwift                      // -20% move cooldown
	AffixLucky                      // +5% crit chance
	AffixExecutioner                // Double damage to enemies below 25% HP

	// Armor affixes
	AffixThorns       // Reflect 2 damage to melee attackers
	AffixRegeneration // +1 HP every 3 seconds
	AffixFortified    // +25% max HP
	AffixEvasion      // +5% dodge chance
	AffixFireproof    // Immune to fire/lava hazards
	AffixAbsorbing    // 10% chance to heal instead of take damage
	AffixStealth      // Enemy detection range -2
	AffixBulwark      // Reduce all damage by 1

	affixCount // sentinel for total count
)

// Affix holds display info for an affix type.
type Affix struct {
	ID   AffixID
	Name string
	Desc string
}

var weaponAffixes = []AffixID{
	AffixVampiric, AffixBurning, AffixFreezing, AffixVenomous,
	AffixBerserker, AffixSwift, AffixLucky, AffixExecutioner,
}

var armorAffixes = []AffixID{
	AffixThorns, AffixRegeneration, AffixFortified, AffixEvasion,
	AffixFireproof, AffixAbsorbing, AffixStealth, AffixBulwark,
}

// AffixDefs maps affix IDs to their display data.
var AffixDefs = map[AffixID]Affix{
	AffixVampiric:     {AffixVampiric, "Vampiric", "Heal 1 HP on kill"},
	AffixBurning:      {AffixBurning, "Burning", "+3 fire damage"},
	AffixFreezing:     {AffixFreezing, "Freezing", "15% chance to stun"},
	AffixVenomous:     {AffixVenomous, "Venomous", "20% chance to poison"},
	AffixBerserker:    {AffixBerserker, "Berserker", "+50% dmg below 30% HP"},
	AffixSwift:        {AffixSwift, "Swift", "-20% move cooldown"},
	AffixLucky:        {AffixLucky, "Lucky", "+5% crit chance"},
	AffixExecutioner:  {AffixExecutioner, "Executioner", "2x dmg to low HP foes"},
	AffixThorns:       {AffixThorns, "Thorns", "Reflect 2 damage"},
	AffixRegeneration: {AffixRegeneration, "Regeneration", "+1 HP / 3s"},
	AffixFortified:    {AffixFortified, "Fortified", "+25% max HP"},
	AffixEvasion:      {AffixEvasion, "Evasion", "+5% dodge"},
	AffixFireproof:    {AffixFireproof, "Fireproof", "Immune to fire"},
	AffixAbsorbing:    {AffixAbsorbing, "Absorbing", "10% heal on hit taken"},
	AffixStealth:      {AffixStealth, "Stealth", "Detection range -2"},
	AffixBulwark:      {AffixBulwark, "Bulwark", "All damage -1"},
}

// RollAffixes generates random affixes for an item based on its rarity and type.
func RollAffixes(rng *rand.Rand, itemType ItemType, rarity Rarity) []AffixID {
	var pool []AffixID
	switch itemType {
	case ItemWeapon:
		pool = weaponAffixes
	case ItemArmor:
		pool = armorAffixes
	default:
		return nil
	}

	count := affixCountForRarity(rarity, rng)
	if count == 0 || len(pool) == 0 {
		return nil
	}

	// Pick unique affixes via partial Fisher-Yates (no full perm allocation)
	indices := make([]int, len(pool))
	for i := range indices {
		indices[i] = i
	}
	result := make([]AffixID, 0, count)
	for i := 0; i < count && i < len(indices); i++ {
		j := i + rng.Intn(len(indices)-i)
		indices[i], indices[j] = indices[j], indices[i]
		result = append(result, pool[indices[i]])
	}
	return result
}

func affixCountForRarity(rarity Rarity, rng *rand.Rand) int {
	switch rarity {
	case RarityCommon:
		return 0
	case RarityUncommon:
		if rng.Intn(2) == 0 {
			return 1
		}
		return 0
	case RarityRare:
		return 1
	case RarityEpic:
		if rng.Intn(2) == 0 {
			return 2
		}
		return 1
	}
	return 0
}

// HasAffix returns true if the affix list contains the given affix.
func HasAffix(affixes []AffixID, id AffixID) bool {
	for _, a := range affixes {
		if a == id {
			return true
		}
	}
	return false
}

// AffixNames returns a comma-separated string of affix names.
func AffixNames(affixes []AffixID) string {
	if len(affixes) == 0 {
		return ""
	}
	s := ""
	for i, id := range affixes {
		if i > 0 {
			s += ", "
		}
		if def, ok := AffixDefs[id]; ok {
			s += def.Name
		}
	}
	return s
}
