package game

import "math/rand"

// AIState controls enemy behavior.
type AIState int

const (
	AIIdle AIState = iota
	AIPatrol
	AIChase
	AIFlee
)

// EnemyType identifies the kind of enemy.
type EnemyType int

const (
	EnemyRat EnemyType = iota
	EnemySkeleton
	EnemyBat
	EnemySlime
	EnemyGhost
	// Bosses
	EnemyMinotaur
	EnemyLich
	EnemyDragon
)

// EnemyDef defines base stats and properties for an enemy type.
type EnemyDef struct {
	Type   EnemyType
	Sprite string
	Stats  Stats
	BaseXP int
}

var EnemyDefs = map[EnemyType]EnemyDef{
	EnemyRat: {
		Type:   EnemyRat,
		Sprite: "rat",
		Stats:  Stats{STR: 1, DEX: 4, VIT: 1, LCK: 1, Level: 1},
		BaseXP: 10,
	},
	EnemySkeleton: {
		Type:   EnemySkeleton,
		Sprite: "skeleton",
		Stats:  Stats{STR: 3, DEX: 2, VIT: 3, LCK: 1, Level: 2},
		BaseXP: 25,
	},
	EnemyBat: {
		Type:   EnemyBat,
		Sprite: "bat",
		Stats:  Stats{STR: 1, DEX: 6, VIT: 1, LCK: 2, Level: 1},
		BaseXP: 15,
	},
	EnemySlime: {
		Type:   EnemySlime,
		Sprite: "slime",
		Stats:  Stats{STR: 2, DEX: 1, VIT: 4, LCK: 1, Level: 2},
		BaseXP: 20,
	},
	EnemyGhost: {
		Type:   EnemyGhost,
		Sprite: "ghost",
		Stats:  Stats{STR: 3, DEX: 3, VIT: 2, LCK: 3, Level: 3},
		BaseXP: 35,
	},
	EnemyMinotaur: {
		Type:   EnemyMinotaur,
		Sprite: "minotaur",
		Stats:  Stats{STR: 8, DEX: 3, VIT: 12, LCK: 2, Level: 5},
		BaseXP: 100,
	},
	EnemyLich: {
		Type:   EnemyLich,
		Sprite: "lich",
		Stats:  Stats{STR: 10, DEX: 5, VIT: 10, LCK: 5, Level: 8},
		BaseXP: 200,
	},
	EnemyDragon: {
		Type:   EnemyDragon,
		Sprite: "dragon",
		Stats:  Stats{STR: 15, DEX: 4, VIT: 20, LCK: 6, Level: 12},
		BaseXP: 500,
	},
}

// BossForFloor returns the boss type for a given floor, or -1 if not a boss floor.
func BossForFloor(floor int) EnemyType {
	switch {
	case floor%15 == 0:
		return EnemyDragon
	case floor%10 == 0:
		return EnemyLich
	case floor%5 == 0:
		return EnemyMinotaur
	}
	return -1
}

// IsBoss returns true if this enemy type is a boss.
func (e EnemyType) IsBoss() bool {
	return e == EnemyMinotaur || e == EnemyLich || e == EnemyDragon
}

// Spawn tables: which enemies can appear at which floor depths.
// Each entry is (EnemyType, weight). Higher weight = more likely.
var SpawnTable = []struct {
	MinFloor int
	Type     EnemyType
	Weight   int
}{
	{1, EnemyRat, 10},
	{1, EnemyBat, 5},
	{2, EnemySlime, 8},
	{3, EnemySkeleton, 7},
	{5, EnemyGhost, 4},
}

// Enemy is a hostile entity with AI behavior.
type Enemy struct {
	*Entity
	Type        EnemyType
	AI          AIState
	ActionTimer float64
	PatrolDX    int
	PatrolDY    int
	PatrolSteps int
	BaseXP      int
}

// NewEnemy creates an enemy from a definition, scaled to the given floor.
func NewEnemy(def EnemyDef, x, y, floor int) *Enemy {
	stats := def.Stats
	stats.Level = def.Stats.Level + floor/3

	// Scale HP with floor
	hpMult := 1.0
	if floor >= 5 {
		hpMult = 1.5
	}
	if floor >= 10 {
		hpMult = 2.0
	}
	if floor >= 15 {
		hpMult = 3.0
	}
	stats.VIT = int(float64(stats.VIT) * hpMult)
	stats.HP = stats.MaxHP()

	return &Enemy{
		Entity:      NewEntity(x, y, def.Sprite, stats),
		Type:        def.Type,
		AI:          AIPatrol,
		ActionTimer: 0,
		BaseXP:      def.BaseXP,
	}
}

// PickEnemyType selects a random enemy type appropriate for the given floor.
func PickEnemyType(floor int, rng *rand.Rand) EnemyType {
	totalWeight := 0
	for _, entry := range SpawnTable {
		if floor >= entry.MinFloor {
			totalWeight += entry.Weight
		}
	}
	if totalWeight == 0 {
		return EnemyRat
	}

	roll := rng.Intn(totalWeight)
	for _, entry := range SpawnTable {
		if floor >= entry.MinFloor {
			roll -= entry.Weight
			if roll < 0 {
				return entry.Type
			}
		}
	}
	return EnemyRat
}

// Update ticks enemy timers and runs AI.
func (e *Enemy) Update(dt float64, playerX, playerY int, passable func(x, y int) bool) {
	e.ActionTimer -= dt
	if e.ActionTimer > 0 {
		return
	}

	// Reset action cooldown based on DEX
	e.ActionTimer = e.Stats.MoveCooldownMS() / 1000.0

	switch e.AI {
	case AIPatrol:
		e.updatePatrol(playerX, playerY, passable)
	case AIChase:
		e.updateChase(playerX, playerY, passable)
	case AIFlee:
		e.updateFlee(playerX, playerY, passable)
	case AIIdle:
		// Check if player is nearby to switch to chase
		if e.distTo(playerX, playerY) <= 6 {
			e.AI = AIChase
		}
	}
}

func (e *Enemy) updatePatrol(playerX, playerY int, passable func(x, y int) bool) {
	// Check for player in detection range
	if e.distTo(playerX, playerY) <= 6 {
		e.AI = AIChase
		return
	}

	// Wander randomly
	if e.PatrolSteps <= 0 {
		dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
		d := dirs[rand.Intn(len(dirs))]
		e.PatrolDX = d[0]
		e.PatrolDY = d[1]
		e.PatrolSteps = 2 + rand.Intn(4)
	}

	nx, ny := e.X+e.PatrolDX, e.Y+e.PatrolDY
	if passable(nx, ny) {
		e.X = nx
		e.Y = ny
		e.PatrolSteps--
	} else {
		e.PatrolSteps = 0 // pick new direction next tick
	}
}

func (e *Enemy) updateChase(playerX, playerY int, passable func(x, y int) bool) {
	// Flee if low HP
	maxHP := e.Stats.MaxHP()
	if maxHP > 0 && float64(e.Stats.HP)/float64(maxHP) < 0.25 {
		e.AI = AIFlee
		return
	}

	// Lose interest if player is far
	if e.distTo(playerX, playerY) > 10 {
		e.AI = AIPatrol
		return
	}

	// Move toward player (simple direct-line)
	dx, dy := sign(playerX-e.X), sign(playerY-e.Y)

	// Try horizontal + vertical, then just horizontal, then just vertical
	if dx != 0 && dy != 0 {
		if passable(e.X+dx, e.Y+dy) && !(e.X+dx == playerX && e.Y+dy == playerY) {
			e.X += dx
			e.Y += dy
			return
		}
	}
	if dx != 0 && passable(e.X+dx, e.Y) && !(e.X+dx == playerX && e.Y == playerY) {
		e.X += dx
		return
	}
	if dy != 0 && passable(e.X, e.Y+dy) && !(e.X == playerX && e.Y+dy == playerY) {
		e.Y += dy
		return
	}
}

func (e *Enemy) updateFlee(playerX, playerY int, passable func(x, y int) bool) {
	// If HP recovered, go back to patrol
	maxHP := e.Stats.MaxHP()
	if maxHP > 0 && float64(e.Stats.HP)/float64(maxHP) >= 0.5 {
		e.AI = AIPatrol
		return
	}

	// Move away from player
	dx, dy := sign(e.X-playerX), sign(e.Y-playerY)
	if dx != 0 && passable(e.X+dx, e.Y) {
		e.X += dx
		return
	}
	if dy != 0 && passable(e.X, e.Y+dy) {
		e.Y += dy
		return
	}
}

func (e *Enemy) distTo(x, y int) int {
	dx := e.X - x
	dy := e.Y - y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}

func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}
