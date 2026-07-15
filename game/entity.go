package game

// Entity is the base for all game entities (player, enemies, NPCs).
type Entity struct {
	X, Y    int
	Stats   Stats
	Sprite  string // sprite key used for rendering
	IsAlive bool
}

// NewEntity creates a living entity at the given position.
func NewEntity(x, y int, sprite string, stats Stats) *Entity {
	stats.HP = stats.MaxHP()
	return &Entity{
		X:       x,
		Y:       y,
		Stats:   stats,
		Sprite:  sprite,
		IsAlive: true,
	}
}

// TakeDamage reduces HP and marks entity dead if HP <= 0. Returns actual damage dealt.
func (e *Entity) TakeDamage(amount int) int {
	if amount < 0 {
		amount = 0
	}
	if amount > e.Stats.HP {
		amount = e.Stats.HP
	}
	e.Stats.HP -= amount
	if e.Stats.HP <= 0 {
		e.Stats.HP = 0
		e.IsAlive = false
	}
	return amount
}

// Heal restores HP up to max. Returns actual amount healed.
func (e *Entity) Heal(amount int) int {
	max := e.Stats.MaxHP()
	before := e.Stats.HP
	e.Stats.HP += amount
	if e.Stats.HP > max {
		e.Stats.HP = max
	}
	return e.Stats.HP - before
}
