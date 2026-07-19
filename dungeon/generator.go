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
	Map           *DungeonMap
	Rooms         []*Room
	SpawnX        int
	SpawnY        int
	StairsX       int
	StairsY       int
	SecretRoomIdx int // index of secret room in Rooms, or -1 if none
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

	// Place cracked (destructible) walls
	placeCrackedWalls(dm, rooms, rng)

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

	// Attempt to place a secret room (disconnected, accessible via cracked wall)
	secretIdx := -1
	if secretRoom := placeSecretRoom(dm, rooms, rng); secretRoom != nil {
		secretIdx = len(rooms)
		rooms = append(rooms, secretRoom)
	}

	return &GenerateResult{
		Map:           dm,
		Rooms:         rooms,
		SpawnX:        spawnX,
		SpawnY:        spawnY,
		StairsX:       stairsX,
		StairsY:       stairsY,
		SecretRoomIdx: secretIdx,
	}
}

// placeSecretRoom tries to carve a small hidden room adjacent to an existing room,
// connected only by cracked walls. Returns nil if placement fails.
func placeSecretRoom(dm *DungeonMap, rooms []*Room, rng *rand.Rand) *Room {
	// Try up to 20 times to find a valid placement
	for attempt := 0; attempt < 20; attempt++ {
		// Pick a random existing room to attach to
		srcIdx := rng.Intn(len(rooms))
		src := rooms[srcIdx]

		// Pick a random side: 0=top, 1=bottom, 2=left, 3=right
		side := rng.Intn(4)
		sw, sh := 4+rng.Intn(3), 4+rng.Intn(2) // secret room size 4-6 x 4-5

		var sx, sy int // top-left of secret room
		var cx, cy int // cracked wall position

		switch side {
		case 0: // above
			sx = src.X + rng.Intn(max(1, src.W-sw))
			sy = src.Y - sh - 1
			cx = sx + sw/2
			cy = src.Y - 1
		case 1: // below
			sx = src.X + rng.Intn(max(1, src.W-sw))
			sy = src.Y + src.H + 1
			cx = sx + sw/2
			cy = src.Y + src.H
		case 2: // left
			sx = src.X - sw - 1
			sy = src.Y + rng.Intn(max(1, src.H-sh))
			cx = src.X - 1
			cy = sy + sh/2
		case 3: // right
			sx = src.X + src.W + 1
			sy = src.Y + rng.Intn(max(1, src.H-sh))
			cx = src.X + src.W
			cy = sy + sh/2
		}

		// Validate bounds
		if sx < 1 || sy < 1 || sx+sw >= MapWidth-1 || sy+sh >= MapHeight-1 {
			continue
		}

		// Check the area is all walls (not overlapping other rooms)
		valid := true
		for y := sy - 1; y <= sy+sh; y++ {
			for x := sx - 1; x <= sx+sw; x++ {
				if dm.At(x, y) != TileWall {
					valid = false
					break
				}
			}
			if !valid {
				break
			}
		}
		if !valid {
			continue
		}

		// Carve the secret room
		room := &Room{X: sx, Y: sy, W: sw, H: sh}
		carveRoom(dm, room)

		// Place cracked wall as the only access point
		dm.Set(cx, cy, TileCrackedWall)

		return room
	}
	return nil
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

// placeCrackedWalls randomly converts some walls adjacent to rooms into cracked walls.
func placeCrackedWalls(dm *DungeonMap, rooms []*Room, rng *rand.Rand) {
	for _, r := range rooms {
		// Check perimeter walls of each room
		for x := r.X - 1; x <= r.X+r.W; x++ {
			maybeCrack(dm, x, r.Y-1, rng)
			maybeCrack(dm, x, r.Y+r.H, rng)
		}
		for y := r.Y; y < r.Y+r.H; y++ {
			maybeCrack(dm, r.X-1, y, rng)
			maybeCrack(dm, r.X+r.W, y, rng)
		}
	}
}

func maybeCrack(dm *DungeonMap, x, y int, rng *rand.Rand) {
	if !dm.InBounds(x, y) || dm.At(x, y) != TileWall {
		return
	}
	// 5% chance to become a cracked wall
	if rng.Intn(100) < 5 {
		dm.Set(x, y, TileCrackedWall)
	}
}
