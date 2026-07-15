package game

// KillStreak tracks rapid kills for combo bonuses.
type KillStreak struct {
	Count    int
	Timer    float64 // seconds since last kill
	Timeout  float64 // seconds before streak resets
	BonusXP  float64 // XP multiplier for current streak
}

// NewKillStreak creates a kill streak tracker.
func NewKillStreak() *KillStreak {
	return &KillStreak{
		Timeout: 3.0, // 3 seconds to chain kills
	}
}

// RegisterKill records a kill and returns the XP multiplier.
func (ks *KillStreak) RegisterKill() float64 {
	ks.Count++
	ks.Timer = 0

	// XP multiplier: 1.0x base, +0.25x per streak kill (max 3.0x)
	ks.BonusXP = 1.0 + float64(ks.Count-1)*0.25
	if ks.BonusXP > 3.0 {
		ks.BonusXP = 3.0
	}
	return ks.BonusXP
}

// Update ticks the streak timer. Resets streak if timeout expires.
func (ks *KillStreak) Update(dt float64) {
	if ks.Count == 0 {
		return
	}
	ks.Timer += dt
	if ks.Timer >= ks.Timeout {
		ks.Count = 0
		ks.BonusXP = 1.0
		ks.Timer = 0
	}
}

// Active returns true if a streak is in progress.
func (ks *KillStreak) Active() bool {
	return ks.Count >= 2
}

// Label returns a display string for the current streak.
func (ks *KillStreak) Label() string {
	switch {
	case ks.Count >= 8:
		return "UNSTOPPABLE!"
	case ks.Count >= 5:
		return "RAMPAGE!"
	case ks.Count >= 3:
		return "KILLING SPREE!"
	case ks.Count >= 2:
		return "DOUBLE KILL!"
	}
	return ""
}
