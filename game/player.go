package game

// Player wraps an Entity with player-specific state.
type Player struct {
	*Entity
	MoveCooldown float64 // seconds remaining until next move
	LastCombat   float64 // seconds since last combat (for HP regen)
	RegenTimer   float64 // accumulator for passive regen ticks
}

// NewPlayer creates a player entity at the given position.
func NewPlayer(x, y int) *Player {
	stats := DefaultPlayerStats()
	return &Player{
		Entity:     NewEntity(x, y, "player", stats),
		LastCombat: 999, // start with regen available
	}
}

// CanMove returns true if the movement cooldown has elapsed.
func (p *Player) CanMove() bool {
	return p.MoveCooldown <= 0
}

// ResetMoveCooldown sets the cooldown based on DEX.
func (p *Player) ResetMoveCooldown() {
	p.MoveCooldown = p.Stats.MoveCooldownMS() / 1000.0
}

// Update ticks player timers.
func (p *Player) Update(dt float64) {
	if p.MoveCooldown > 0 {
		p.MoveCooldown -= dt
	}
	p.LastCombat += dt
}
