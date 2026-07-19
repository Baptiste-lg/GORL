package game

import (
	"math/rand"

	"github.com/Baptiste-lg/GORL/dungeon"
)

// World holds all entities and the dungeon for the current floor.
type World struct {
	Dungeon *dungeon.GenerateResult
	Enemies []*Enemy
	Traps   []*Trap
	Hazards []*Hazard
	Shrines []*Shrine
	Floor   int
}

// NewWorld creates a world with enemies, traps, hazards, and shrines.
func NewWorld(dg *dungeon.GenerateResult, floor int, rng *rand.Rand) *World {
	// Convert dungeon rooms to game rooms for trap/hazard/shrine spawning
	rooms := make([]*Room, len(dg.Rooms))
	for i, r := range dg.Rooms {
		rooms[i] = &Room{X: r.X, Y: r.Y, W: r.W, H: r.H}
	}

	w := &World{
		Dungeon: dg,
		Floor:   floor,
		Traps:   SpawnTraps(rooms, floor, rng),
		Hazards: SpawnHazards(rooms, floor, rng),
		Shrines: SpawnShrines(rooms, floor, rng),
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
		if i == 0 {
			continue
		}

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

