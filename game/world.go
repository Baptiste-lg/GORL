package game

import (
	"math/rand"

	"github.com/Baptiste-lg/GORL/dungeon"
)

// World holds all entities and the dungeon for the current floor.
type World struct {
	Dungeon *dungeon.GenerateResult
	Enemies []*Enemy
	Floor   int
}

// NewWorld creates a world with enemies spawned in rooms.
func NewWorld(dg *dungeon.GenerateResult, floor int, rng *rand.Rand) *World {
	w := &World{
		Dungeon: dg,
		Floor:   floor,
	}
	w.spawnEnemies(rng)
	return w
}

func (w *World) spawnEnemies(rng *rand.Rand) {
	for i, room := range w.Dungeon.Rooms {
		// Skip the player's spawn room
		if i == 0 {
			continue
		}

		// 1-3 enemies per room, scaling with floor
		count := 1 + rng.Intn(3)
		if w.Floor >= 5 {
			count += 1
		}
		if w.Floor >= 10 {
			count += 1
		}

		for j := 0; j < count; j++ {
			eType := PickEnemyType(w.Floor, rng)
			def := EnemyDefs[eType]

			// Random position within the room
			x := room.X + 1 + rng.Intn(max(1, room.W-2))
			y := room.Y + 1 + rng.Intn(max(1, room.H-2))

			// Don't spawn on stairs
			tile := w.Dungeon.Map.At(x, y)
			if tile == dungeon.TileStairsDown || tile == dungeon.TileStairsUp {
				continue
			}

			// Don't stack enemies
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

			w.Enemies = append(w.Enemies, NewEnemy(def, x, y, w.Floor))
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
			// Don't walk into other enemies
			for _, other := range w.Enemies {
				if other != e && other.IsAlive && other.X == x && other.Y == y {
					return false
				}
			}
			return true
		})
	}
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

// RemoveDead filters out dead enemies.
func (w *World) RemoveDead() []*Enemy {
	var dead []*Enemy
	alive := w.Enemies[:0]
	for _, e := range w.Enemies {
		if e.IsAlive {
			alive = append(alive, e)
		} else {
			dead = append(dead, e)
		}
	}
	w.Enemies = alive
	return dead
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
