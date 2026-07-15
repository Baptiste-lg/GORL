package dungeon

import "math/rand"

// carveCorridors connects rooms with L-shaped corridors.
func carveCorridors(dm *DungeonMap, rooms []*Room, rng *rand.Rand) {
	for i := 1; i < len(rooms); i++ {
		x1, y1 := rooms[i-1].CenterX(), rooms[i-1].CenterY()
		x2, y2 := rooms[i].CenterX(), rooms[i].CenterY()

		if rng.Intn(2) == 0 {
			// Horizontal first, then vertical
			carveHLine(dm, x1, x2, y1)
			carveVLine(dm, y1, y2, x2)
		} else {
			// Vertical first, then horizontal
			carveVLine(dm, y1, y2, x1)
			carveHLine(dm, x1, x2, y2)
		}
	}
}

// carveHLine carves a horizontal corridor from x1 to x2 at row y.
func carveHLine(dm *DungeonMap, x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if dm.InBounds(x, y) && dm.At(x, y) == TileWall {
			dm.Set(x, y, TileFloor)
		}
	}
}

// carveVLine carves a vertical corridor from y1 to y2 at column x.
func carveVLine(dm *DungeonMap, y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if dm.InBounds(x, y) && dm.At(x, y) == TileWall {
			dm.Set(x, y, TileFloor)
		}
	}
}
