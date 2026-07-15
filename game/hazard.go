package game

import "math/rand"

// HazardType identifies the kind of environmental hazard.
type HazardType int

const (
	HazardLava HazardType = iota
	HazardPoisonGas
)

// Hazard is a persistent area-of-effect zone on the map.
type Hazard struct {
	X, Y     int
	Type     HazardType
	DPS      float64 // damage per second
	DmgAccum float64 // accumulated fractional damage
}

// Glyph returns the display character.
func (h *Hazard) Glyph() string {
	switch h.Type {
	case HazardLava:
		return "~"
	case HazardPoisonGas:
		return "%"
	}
	return "~"
}

// Color returns the display color.
func (h *Hazard) Color() string {
	switch h.Type {
	case HazardLava:
		return "#ff4400"
	case HazardPoisonGas:
		return "#44cc44"
	}
	return "#ff4400"
}

// SpawnHazards places environmental hazards in rooms.
func SpawnHazards(rooms []*Room, floor int, rng *rand.Rand) []*Hazard {
	var hazards []*Hazard

	// No hazards on first few floors
	if floor < 3 {
		return nil
	}

	for i, room := range rooms {
		if i == 0 {
			continue
		}

		// Small chance of a hazard cluster per room
		if rng.Intn(100) > 30 {
			continue
		}

		hazType := HazardLava
		dps := 3.0
		if floor >= 5 && rng.Intn(2) == 0 {
			hazType = HazardPoisonGas
			dps = 2.0
		}

		// Place a small cluster (2-4 tiles)
		cx := room.X + 1 + rng.Intn(maxInt(1, room.W-2))
		cy := room.Y + 1 + rng.Intn(maxInt(1, room.H-2))
		clusterSize := 2 + rng.Intn(3)

		for k := 0; k < clusterSize; k++ {
			hx := cx + rng.Intn(3) - 1
			hy := cy + rng.Intn(3) - 1
			hazards = append(hazards, &Hazard{
				X:    hx,
				Y:    hy,
				Type: hazType,
				DPS:  dps,
			})
		}
	}

	return hazards
}
