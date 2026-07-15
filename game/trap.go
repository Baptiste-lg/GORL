package game

import "math/rand"

// TrapType identifies the kind of trap.
type TrapType int

const (
	TrapSpike TrapType = iota
	TrapPoison
	TrapTeleport
)

// Trap is a hidden hazard on a floor tile.
type Trap struct {
	X, Y      int
	Type      TrapType
	Revealed  bool
	Triggered bool
}

// Glyph returns the display character for a revealed trap.
func (t *Trap) Glyph() string {
	switch t.Type {
	case TrapSpike:
		return "^"
	case TrapPoison:
		return "~"
	case TrapTeleport:
		return "*"
	}
	return "^"
}

// Color returns the display color for a revealed trap.
func (t *Trap) Color() string {
	switch t.Type {
	case TrapSpike:
		return "#aa4444"
	case TrapPoison:
		return "#44aa44"
	case TrapTeleport:
		return "#4444ee"
	}
	return "#aa4444"
}

// SpawnTraps places hidden traps in dungeon rooms.
func SpawnTraps(rooms []*Room, floor int, rng *rand.Rand) []*Trap {
	var traps []*Trap

	for i, room := range rooms {
		// Skip spawn room
		if i == 0 {
			continue
		}

		// 0-2 traps per room, more on deeper floors
		count := rng.Intn(2)
		if floor >= 5 {
			count += rng.Intn(2)
		}

		for j := 0; j < count; j++ {
			x := room.X + 1 + rng.Intn(maxInt(1, room.W-2))
			y := room.Y + 1 + rng.Intn(maxInt(1, room.H-2))

			trapType := TrapType(rng.Intn(3))
			traps = append(traps, &Trap{
				X:    x,
				Y:    y,
				Type: trapType,
			})
		}
	}

	return traps
}

// Room is imported from dungeon but we need it here for spawn.
// Re-use the dungeon.Room type via the world layer.
type Room struct {
	X, Y, W, H int
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
