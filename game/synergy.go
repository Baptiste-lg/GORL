package game

// SynergyID identifies a specific synergy.
type SynergyID int

const (
	// Weapon + Weapon affix combos
	SynToxicFlame     SynergyID = iota // Burning + Venomous
	SynShatter                          // Freezing + Executioner
	SynDesperateGambit                  // Lucky + Berserker
	SynWhirlwind                        // Swift + Lucky
	SynBloodthirst                      // Vampiric + Berserker

	// Weapon + Armor affix combos
	SynFlameLord      // Burning + Fireproof
	SynToxicResilience // Venomous + Regeneration
	SynBloodMirror    // Vampiric + Thorns
	SynIceDancer      // Freezing + Evasion
	SynShadowStep     // Swift + Stealth
	SynJuggernaut     // Berserker + Fortified
	SynReaperGuard    // Executioner + Bulwark
	SynFortuneFavor   // Lucky + Absorbing

	// Active + Affix combos
	SynInferno        // Fireball + Burning
	SynAbsoluteZero   // Freeze + Freezing
	SynCrimsonTide    // HealBurst + Vampiric
	SynFountainOfLife // HealBurst + Regeneration
	SynUnbreakable    // ShieldWall + Fortified
	SynIronMaiden     // ShieldWall + Thorns
	SynPlagueBearer   // PoisonCloud + Venomous
	SynRageUnleashed  // WarCry + Berserker
	SynFlashStep      // Dash + Swift
	SynVanish         // Blink + Stealth
	SynPhantom        // Dash + Evasion

	synergyCount
)

// Synergy defines what's needed for a synergy to activate and what it does.
type Synergy struct {
	ID       SynergyID
	Name     string
	Desc     string
	Affixes  []AffixID // required affixes (across all equipment)
	ActiveReq ActiveID  // required active item (-1 for none)
}

// AllSynergies defines every synergy in the game.
var AllSynergies = []Synergy{
	// --- Weapon + Weapon ---
	{SynToxicFlame, "Toxic Flame", "Fire and poison both proc, +2 each",
		[]AffixID{AffixBurning, AffixVenomous}, ActiveNone},
	{SynShatter, "Shatter", "Frozen foes take 3x crit damage",
		[]AffixID{AffixFreezing, AffixExecutioner}, ActiveNone},
	{SynDesperateGambit, "Desperate Gambit", "Crit +15% below 30% HP",
		[]AffixID{AffixLucky, AffixBerserker}, ActiveNone},
	{SynWhirlwind, "Whirlwind", "Crits halve next move cooldown",
		[]AffixID{AffixSwift, AffixLucky}, ActiveNone},
	{SynBloodthirst, "Bloodthirst", "Heal 3 on kill below 30% HP",
		[]AffixID{AffixVampiric, AffixBerserker}, ActiveNone},

	// --- Weapon + Armor ---
	{SynFlameLord, "Flame Lord", "Burning +5 damage, immune to fire",
		[]AffixID{AffixBurning, AffixFireproof}, ActiveNone},
	{SynToxicResilience, "Toxic Resilience", "Regen heals +2 per tick",
		[]AffixID{AffixVenomous, AffixRegeneration}, ActiveNone},
	{SynBloodMirror, "Blood Mirror", "Thorns also heal you",
		[]AffixID{AffixVampiric, AffixThorns}, ActiveNone},
	{SynIceDancer, "Ice Dancer", "Dodge +10%",
		[]AffixID{AffixFreezing, AffixEvasion}, ActiveNone},
	{SynShadowStep, "Shadow Step", "Detection range 3 tiles",
		[]AffixID{AffixSwift, AffixStealth}, ActiveNone},
	{SynJuggernaut, "Juggernaut", "+50% HP and +50% dmg at low HP",
		[]AffixID{AffixBerserker, AffixFortified}, ActiveNone},
	{SynReaperGuard, "Reaper's Guard", "Execute + reduce all dmg by 2",
		[]AffixID{AffixExecutioner, AffixBulwark}, ActiveNone},
	{SynFortuneFavor, "Fortune's Favor", "Absorb chance +15%",
		[]AffixID{AffixLucky, AffixAbsorbing}, ActiveNone},

	// --- Active + Affix ---
	{SynInferno, "Inferno", "Fireball: 5 tiles, +10 damage",
		[]AffixID{AffixBurning}, ActiveFireball},
	{SynAbsoluteZero, "Absolute Zero", "Freeze lasts 4s",
		[]AffixID{AffixFreezing}, ActiveFreeze},
	{SynCrimsonTide, "Crimson Tide", "Heal burst +5 HP per visible foe",
		[]AffixID{AffixVampiric}, ActiveHealBurst},
	{SynFountainOfLife, "Fountain of Life", "Regen doubled 10s after heal",
		[]AffixID{AffixRegeneration}, ActiveHealBurst},
	{SynUnbreakable, "Unbreakable", "Shield gives +20 VIT",
		[]AffixID{AffixFortified}, ActiveShieldWall},
	{SynIronMaiden, "Iron Maiden", "Thorns x3 while shielded",
		[]AffixID{AffixThorns}, ActiveShieldWall},
	{SynPlagueBearer, "Plague Bearer", "Poison cloud double radius",
		[]AffixID{AffixVenomous}, ActivePoisonCloud},
	{SynRageUnleashed, "Rage Unleashed", "War cry gives +10 STR",
		[]AffixID{AffixBerserker}, ActiveWarCry},
	{SynFlashStep, "Flash Step", "Dash 5 tiles, -cooldown 2s",
		[]AffixID{AffixSwift}, ActiveDash},
	{SynVanish, "Vanish", "After blink, enemies lose aggro",
		[]AffixID{AffixStealth}, ActiveBlink},
	{SynPhantom, "Phantom", "100% dodge 1s after dash",
		[]AffixID{AffixEvasion}, ActiveDash},
}

// DetectSynergies returns the list of active synergy IDs given the player's equipment and active item.
func DetectSynergies(weapon, armor *Item, active *ActiveItem) []SynergyID {
	// Collect all affixes from weapon + armor
	var allAffixes []AffixID
	if weapon != nil {
		allAffixes = append(allAffixes, weapon.Affixes...)
	}
	if armor != nil {
		allAffixes = append(allAffixes, armor.Affixes...)
	}

	activeID := ActiveNone
	if active != nil {
		activeID = active.ID
	}

	var result []SynergyID
	for _, syn := range AllSynergies {
		if syn.ActiveReq != ActiveNone && syn.ActiveReq != activeID {
			continue
		}
		if hasAllAffixes(allAffixes, syn.Affixes) {
			result = append(result, syn.ID)
		}
	}
	return result
}

func hasAllAffixes(have []AffixID, need []AffixID) bool {
	for _, n := range need {
		found := false
		for _, h := range have {
			if h == n {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// SynergyByID returns the synergy definition, or nil if not found.
func SynergyByID(id SynergyID) *Synergy {
	for i := range AllSynergies {
		if AllSynergies[i].ID == id {
			return &AllSynergies[i]
		}
	}
	return nil
}

// HasSynergy returns true if the given synergy ID is in the active list.
func HasSynergy(synergies []SynergyID, id SynergyID) bool {
	for _, s := range synergies {
		if s == id {
			return true
		}
	}
	return false
}
