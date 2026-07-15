package dungeon

import (
	"math"
	"math/rand"
)

const (
	MapWidth  = 50
	MapHeight = 40
)

// GenerateResult holds the output of dungeon generation.
type GenerateResult struct {
	Map      *DungeonMap
	Rooms    []*Room
	SpawnX   int
	SpawnY   int
	StairsX  int
	StairsY  int
}

// Generate creates a new dungeon floor.
func Generate(seed int64) *GenerateResult {
	rng := rand.New(rand.NewSource(seed))
	dm := NewDungeonMap(MapWidth, MapHeight)

	// Generate rooms via BSP
	rooms := generateBSP(MapWidth, MapHeight, rng)
	if len(rooms) == 0 {
		// Fallback: single room
		rooms = []*Room{{X: 5, Y: 5, W: 10, H: 8}}
	}

	// Carve rooms into the map
	for _, r := range rooms {
		carveRoom(dm, r)
	}

	// Connect rooms with corridors
	carveCorridors(dm, rooms, rng)

	// Place doors at room entrances
	placeDoors(dm, rooms)

	// Player spawns in center of first room
	spawnX, spawnY := rooms[0].CenterX(), rooms[0].CenterY()

	// Stairs go in the room farthest from spawn
	farthest := rooms[0]
	maxDist := 0.0
	for _, r := range rooms[1:] {
		dx := float64(r.CenterX() - spawnX)
		dy := float64(r.CenterY() - spawnY)
		d := math.Sqrt(dx*dx + dy*dy)
		if d > maxDist {
			maxDist = d
			farthest = r
		}
	}
	stairsX, stairsY := farthest.CenterX(), farthest.CenterY()
	dm.Set(stairsX, stairsY, TileStairsDown)

	// Mark spawn as stairs up (previous floor)
	dm.Set(spawnX, spawnY, TileStairsUp)

	return &GenerateResult{
		Map:     dm,
		Rooms:   rooms,
		SpawnX:  spawnX,
		SpawnY:  spawnY,
		StairsX: stairsX,
		StairsY: stairsY,
	}
}

// carveRoom sets all tiles inside a room to floor.
func carveRoom(dm *DungeonMap, r *Room) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			dm.Set(x, y, TileFloor)
		}
	}
}

// placeDoors puts doors where corridors meet room edges.
func placeDoors(dm *DungeonMap, rooms []*Room) {
	for _, r := range rooms {
		// Check perimeter of each room
		for x := r.X - 1; x <= r.X+r.W; x++ {
			checkDoor(dm, x, r.Y-1)
			checkDoor(dm, x, r.Y+r.H)
		}
		for y := r.Y; y < r.Y+r.H; y++ {
			checkDoor(dm, r.X-1, y)
			checkDoor(dm, r.X+r.W, y)
		}
	}
}

// checkDoor places a door if a floor tile sits on a room boundary
// with wall neighbors on both sides perpendicular to the corridor.
func checkDoor(dm *DungeonMap, x, y int) {
	if !dm.InBounds(x, y) || dm.At(x, y) != TileFloor {
		return
	}

	// Horizontal passage: walls above and below, floor left and right
	hWalls := dm.At(x, y-1).BlocksSight() && dm.At(x, y+1).BlocksSight()
	hFloor := dm.At(x-1, y).Passable() && dm.At(x+1, y).Passable()

	// Vertical passage: walls left and right, floor above and below
	vWalls := dm.At(x-1, y).BlocksSight() && dm.At(x+1, y).BlocksSight()
	vFloor := dm.At(x, y-1).Passable() && dm.At(x, y+1).Passable()

	if (hWalls && hFloor) || (vWalls && vFloor) {
		dm.Set(x, y, TileDoor)
	}
}
