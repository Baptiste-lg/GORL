package game

// Stats holds the core RPG attributes for an entity.
type Stats struct {
	STR   int // Strength: melee damage
	DEX   int // Dexterity: speed, dodge chance
	VIT   int // Vitality: max HP
	LCK   int // Luck: crit chance, loot quality
	Level int
	HP    int // Current HP (derived from VIT at init)
	XP    int // Current XP (player only)
}

// MaxHP returns the maximum hit points based on VIT.
// Formula: 10 + VIT * 5
func (s Stats) MaxHP() int {
	return 10 + s.VIT*5
}

// Damage returns the base melee damage.
// Formula: STR * 2, minimum 1
func (s Stats) Damage() int {
	d := s.STR * 2
	if d < 1 {
		d = 1
	}
	return d
}

// Defense returns damage reduction.
// Formula: VIT / 2
func (s Stats) Defense() int {
	return s.VIT / 2
}

// MoveCooldownMS returns movement cooldown in milliseconds based on DEX.
// Formula: 200 - DEX * 5, clamped to [80, 200]
func (s Stats) MoveCooldownMS() float64 {
	cd := 200 - s.DEX*5
	if cd < 80 {
		cd = 80
	}
	if cd > 200 {
		cd = 200
	}
	return float64(cd)
}

// CritChance returns critical hit chance as a fraction [0, 0.30].
// Formula: LCK * 2%, max 30%
func (s Stats) CritChance() float64 {
	c := float64(s.LCK) * 0.02
	if c > 0.30 {
		c = 0.30
	}
	return c
}

// DodgeChance returns dodge chance as a fraction [0, 0.25].
// Formula: DEX * 1.5%, max 25%
func (s Stats) DodgeChance() float64 {
	d := float64(s.DEX) * 0.015
	if d > 0.25 {
		d = 0.25
	}
	return d
}

// XPToNextLevel returns XP needed for the next level.
// Formula: 100 * level^1.5
func (s Stats) XPToNextLevel() int {
	if s.Level <= 0 {
		return 100
	}
	// Approximate level^1.5 using integer math
	l := float64(s.Level)
	return int(100.0 * l * sqrt(l))
}

// sqrt is a simple integer-friendly square root.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// CanLevelUp returns true if the entity has enough XP.
func (s Stats) CanLevelUp() bool {
	return s.XP >= s.XPToNextLevel()
}

// LevelUpChoices returns 3 random stat boost options.
type StatBoost struct {
	Stat  string // "STR", "DEX", "VIT", "LCK"
	Label string
	Desc  string
	Apply func(s *Stats)
}

var allBoosts = []StatBoost{
	{"STR", "+2 STR", "Increase melee damage", func(s *Stats) { s.STR += 2 }},
	{"DEX", "+2 DEX", "Move and attack faster", func(s *Stats) { s.DEX += 2 }},
	{"VIT", "+2 VIT", "Gain more HP", func(s *Stats) { s.VIT += 2 }},
	{"LCK", "+2 LCK", "Better crits and loot", func(s *Stats) { s.LCK += 2 }},
}

// DefaultPlayerStats returns starting stats for a new player.
func DefaultPlayerStats() Stats {
	s := Stats{
		STR:   3,
		DEX:   3,
		VIT:   3,
		LCK:   2,
		Level: 1,
	}
	s.HP = s.MaxHP()
	return s
}
