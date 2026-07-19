package game

import (
	"math/rand"

	"github.com/Baptiste-lg/GORL/dungeon"
)

// World holds all entities and the dungeon for the current floor.
type World struct {
	Dungeon          *dungeon.GenerateResult
	Enemies          []*Enemy
	Traps            []*Trap
	Hazards          []*Hazard
	Shrines          []*Shrine
	Shop             *Shop
	Floor            int
	RoomRoles        map[int]RoomRole
	GroundItems      []*Item // items placed by room generation
	ChallengeCleared bool
	ChallengeRoom    int // room index of challenge room (-1 if none)
}

// NewWorld creates a world with enemies, traps, hazards, and shrines.
func NewWorld(dg *dungeon.GenerateResult, floor int, rng *rand.Rand) *World {
	// Convert dungeon rooms to game rooms for trap/hazard/shrine spawning
	rooms := make([]*Room, len(dg.Rooms))
	for i, r := range dg.Rooms {
		rooms[i] = &Room{X: r.X, Y: r.Y, W: r.W, H: r.H}
	}

	w := &World{
		Dungeon:       dg,
		Floor:         floor,
		Traps:         SpawnTraps(rooms, floor, rng),
		Hazards:       SpawnHazards(rooms, floor, rng),
		Shrines:       SpawnShrines(rooms, floor, rng),
		RoomRoles:     make(map[int]RoomRole),
		ChallengeRoom: -1,
	}

	// Assign room roles
	w.RoomRoles[0] = RoleSpawn

	// Find stairs room index
	stairsIdx := 0
	for i, r := range dg.Rooms {
		if r.CenterX() == dg.StairsX && r.CenterY() == dg.StairsY {
			stairsIdx = i
			break
		}
	}
	w.RoomRoles[stairsIdx] = RoleStairs

	// Collect available room indices (not spawn, not stairs)
	var available []int
	for i := range rooms {
		if i == 0 || i == stairsIdx {
			continue
		}
		available = append(available, i)
	}

	// Shuffle available rooms for random assignment
	rng.Shuffle(len(available), func(i, j int) {
		available[i], available[j] = available[j], available[i]
	})

	assigned := 0

	// Shop (non-boss floors)
	if floor%5 != 0 && assigned < len(available) {
		idx := available[assigned]
		assigned++
		w.RoomRoles[idx] = RoleShop
		r := rooms[idx]
		w.Shop = NewShop(r.X+r.W/2, r.Y+r.H/2, floor, rng)
	}

	// Treasure room (always, if available)
	if assigned < len(available) {
		idx := available[assigned]
		assigned++
		w.RoomRoles[idx] = RoleTreasure
		r := rooms[idx]
		loot := GenerateLoot(rng, floor+2) // slightly better loot
		loot.X = r.X + r.W/2
		loot.Y = r.Y + r.H/2
		w.GroundItems = append(w.GroundItems, loot)
	}

	// Challenge room (floor 3+, if available)
	if floor >= 3 && assigned < len(available) {
		idx := available[assigned]
		assigned++
		w.RoomRoles[idx] = RoleChallenge
		w.ChallengeRoom = idx
	}

	// Mark secret room if present (last room if it has no corridor connection)
	// The secret room is always the last room appended by the generator
	// and won't be in our available list since it's unreachable normally.
	// Identify it: any room index not yet assigned a role and not in 'available' remaining
	// Simpler: if floor >= 2, the last room index that has no role is the secret
	lastIdx := len(rooms) - 1
	if floor >= 2 && lastIdx > stairsIdx && w.RoomRoles[lastIdx] == RoleNormal {
		w.RoomRoles[lastIdx] = RoleSecret
		r := rooms[lastIdx]
		secretLoot := GenerateLoot(rng, floor+5)
		secretLoot.X = r.X + r.W/2
		secretLoot.Y = r.Y + r.H/2
		w.GroundItems = append(w.GroundItems, secretLoot)
		// Add bonus gold in secret room
		goldLoot := GenerateLoot(rng, floor+3)
		goldLoot.X = r.X + r.W/2 + 1
		goldLoot.Y = r.Y + r.H/2
		w.GroundItems = append(w.GroundItems, goldLoot)
	}

	w.spawnEnemies(rng)
	return w
}

// ShrineAt returns the shrine at (x, y) or nil.
func (w *World) ShrineAt(x, y int) *Shrine {
	for _, s := range w.Shrines {
		if s.X == x && s.Y == y {
			return s
		}
	}
	return nil
}

func (w *World) spawnEnemies(rng *rand.Rand) {
	// Spawn boss in the stairs room on boss floors
	bossType := BossForFloor(w.Floor)
	if bossType >= 0 {
		def := EnemyDefs[bossType]
		bx := w.Dungeon.StairsX + 1
		by := w.Dungeon.StairsY
		if !w.Dungeon.Map.At(bx, by).Passable() {
			bx = w.Dungeon.StairsX - 1
		}
		boss := NewEnemy(def, bx, by, w.Floor, rng)
		boss.AI = AIChase
		w.Enemies = append(w.Enemies, boss)
	}

	for i, room := range w.Dungeon.Rooms {
		role := w.RoomRoles[i]

		// Skip rooms that shouldn't have enemies
		if role == RoleSpawn || role == RoleShop || role == RoleSecret {
			continue
		}

		count := 1 + rng.Intn(3)
		if w.Floor >= 5 {
			count += 1
		}
		if w.Floor >= 10 {
			count += 1
		}

		// Modify count based on room role
		switch role {
		case RoleTreasure:
			count = rng.Intn(2) // 0-1 enemies
		case RoleChallenge:
			count *= 2 // double enemies
		}

		for j := 0; j < count; j++ {
			eType := PickEnemyType(w.Floor, rng)
			def := EnemyDefs[eType]

			x := room.X + 1 + rng.Intn(max(1, room.W-2))
			y := room.Y + 1 + rng.Intn(max(1, room.H-2))

			tile := w.Dungeon.Map.At(x, y)
			if tile == dungeon.TileStairsDown || tile == dungeon.TileStairsUp {
				continue
			}

			occupied := false
			for _, e := range w.Enemies {
				if e.X == x && e.Y == y {
					occupied = true
					break
				}
			}
			if occupied {
				continue
			}

			w.Enemies = append(w.Enemies, NewEnemy(def, x, y, w.Floor, rng))
		}
	}
}

// UpdateEnemies ticks all living enemies.
func (w *World) UpdateEnemies(dt float64, playerX, playerY int) {
	for _, e := range w.Enemies {
		if !e.IsAlive {
			continue
		}
		e.Update(dt, playerX, playerY, func(x, y int) bool {
			if !w.Dungeon.Map.At(x, y).Passable() {
				return false
			}
			for _, other := range w.Enemies {
				if other != e && other.IsAlive && other.X == x && other.Y == y {
					return false
				}
			}
			return true
		})
	}
}

// TrapAt returns the trap at (x, y) or nil.
func (w *World) TrapAt(x, y int) *Trap {
	for _, t := range w.Traps {
		if t.X == x && t.Y == y && !t.Triggered {
			return t
		}
	}
	return nil
}

// HazardsAt returns all hazards at (x, y).
func (w *World) HazardsAt(x, y int) []*Hazard {
	var result []*Hazard
	for _, h := range w.Hazards {
		if h.X == x && h.Y == y {
			result = append(result, h)
		}
	}
	return result
}

// EnemyAt returns the enemy at (x, y) or nil.
func (w *World) EnemyAt(x, y int) *Enemy {
	for _, e := range w.Enemies {
		if e.IsAlive && e.X == x && e.Y == y {
			return e
		}
	}
	return nil
}

// EnemiesInRoom returns the count of living enemies within a room's bounds.
func (w *World) EnemiesInRoom(roomIdx int) int {
	if roomIdx < 0 || roomIdx >= len(w.Dungeon.Rooms) {
		return 0
	}
	r := w.Dungeon.Rooms[roomIdx]
	count := 0
	for _, e := range w.Enemies {
		if e.IsAlive && e.X >= r.X && e.X < r.X+r.W && e.Y >= r.Y && e.Y < r.Y+r.H {
			count++
		}
	}
	return count
}

// RemoveDead filters out dead enemies.
func (w *World) RemoveDead() {
	alive := w.Enemies[:0]
	for _, e := range w.Enemies {
		if e.IsAlive {
			alive = append(alive, e)
		}
	}
	w.Enemies = alive
}
