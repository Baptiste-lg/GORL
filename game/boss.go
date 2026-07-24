package game

// BossPhase tracks which phase the boss is in based on HP thresholds.
type BossPhase int

const (
	BossPhase1 BossPhase = iota // 100%-50% HP
	BossPhase2                  // 50%-25% HP
	BossPhase3                  // below 25% HP (enraged)
)

// BossState holds boss-specific combat state.
type BossState struct {
	Phase       BossPhase
	AbilityCD   float64 // seconds until next special ability
	ChargeDir   [2]int  // direction of charge attack (Minotaur)
	ChargeDist  int     // tiles remaining in charge
	SummonCount int     // number of summons spawned this phase (Lich)
	BreathDir   [2]int  // fire breath direction (Dragon)
	Enraged     bool    // phase 3 flag
}

// NewBossState creates initial boss state.
func NewBossState() *BossState {
	return &BossState{
		AbilityCD: 2.0, // first ability after 2 seconds
	}
}

// BossCurrentPhase calculates the phase based on HP.
func BossCurrentPhase(hp, maxHP int) BossPhase {
	if maxHP <= 0 {
		return BossPhase1
	}
	ratio := float64(hp) / float64(maxHP)
	switch {
	case ratio <= 0.25:
		return BossPhase3
	case ratio <= 0.50:
		return BossPhase2
	default:
		return BossPhase1
	}
}

// BossAbility represents what a boss does on its special turn.
type BossAbility int

const (
	AbilityNone    BossAbility = iota
	AbilityCharge                     // Minotaur: rush in a line
	AbilityStomp                      // Minotaur: AoE damage around self
	AbilitySummon                     // Lich: spawn skeleton minions
	AbilityDrain                      // Lich: heal from player damage
	AbilityBreath                     // Dragon: fire line AoE
	AbilityRoar                       // Dragon: stun player briefly
)

// DangerTile marks a tile that will be hit next turn (telegraphing).
type DangerTile struct {
	X, Y  int
	Timer float64 // seconds until damage
}
