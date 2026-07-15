package game

import (
	"math/rand"
	"syscall/js"
	"time"

	"github.com/Baptiste-lg/GORL/dungeon"
	"github.com/Baptiste-lg/GORL/render"
)

type State int

const (
	StateMenu State = iota
	StatePlaying
	StatePaused
	StateLevelUp
	StateDead
	StateInventory
)

const (
	fovRadius     = 8
	regenDelay    = 3.0 // seconds after combat before regen starts
	regenInterval = 5.0 // seconds between regen ticks
)

type Game struct {
	state    State
	renderer *render.Renderer
	lastTime float64
	keys     map[string]bool

	dungeonResult *dungeon.GenerateResult
	fov           *dungeon.FOV
	player        *Player
	world         *World
	particles     *render.ParticleSystem
	floor         int
	rng           *rand.Rand
	killCount     int
}

func New() *Game {
	g := &Game{
		state:     StateMenu,
		renderer:  render.NewRenderer(),
		keys:      make(map[string]bool),
		particles: render.NewParticleSystem(),
		floor:     1,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	g.registerInput()
	return g
}

func (g *Game) startNewGame() {
	g.floor = 1
	g.killCount = 0
	g.player = nil
	g.particles = render.NewParticleSystem()
	g.generateFloor()
	g.state = StatePlaying
}

func (g *Game) generateFloor() {
	seed := g.rng.Int63()
	g.dungeonResult = dungeon.Generate(seed)
	g.fov = dungeon.NewFOV(dungeon.MapWidth, dungeon.MapHeight)
	g.world = NewWorld(g.dungeonResult, g.floor, g.rng)

	sx, sy := g.dungeonResult.SpawnX, g.dungeonResult.SpawnY
	if g.player == nil {
		g.player = NewPlayer(sx, sy)
	} else {
		g.player.X = sx
		g.player.Y = sy
	}

	g.renderer.CenterCamera(g.player.X, g.player.Y)
	g.fov.Compute(g.dungeonResult.Map, g.player.X, g.player.Y, fovRadius)
}

func (g *Game) registerInput() {
	doc := js.Global().Get("document")

	doc.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		key := args[0].Get("key").String()
		g.keys[key] = true
		g.handleKeyDown(key)
		args[0].Call("preventDefault")
		return nil
	}))

	doc.Call("addEventListener", "keyup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		key := args[0].Get("key").String()
		g.keys[key] = false
		return nil
	}))
}

func (g *Game) handleKeyDown(key string) {
	switch g.state {
	case StateMenu:
		if key == "Enter" {
			g.startNewGame()
		}
	case StateDead:
		if key == "Enter" {
			g.state = StateMenu
		}
	}
}

func (g *Game) processMovement() {
	if g.player == nil || !g.player.CanMove() {
		return
	}

	dx, dy := 0, 0
	if g.keys["w"] || g.keys["ArrowUp"] {
		dy = -1
	} else if g.keys["s"] || g.keys["ArrowDown"] {
		dy = 1
	}
	if g.keys["a"] || g.keys["ArrowLeft"] {
		dx = -1
	} else if g.keys["d"] || g.keys["ArrowRight"] {
		dx = 1
	}

	if dx == 0 && dy == 0 {
		return
	}

	nx, ny := g.player.X+dx, g.player.Y+dy
	dm := g.dungeonResult.Map
	if !dm.At(nx, ny).Passable() {
		return
	}

	// Attack enemy if one is in the way
	enemy := g.world.EnemyAt(nx, ny)
	if enemy != nil {
		g.playerAttack(enemy)
		g.player.ResetMoveCooldown()
		return
	}

	g.player.X = nx
	g.player.Y = ny
	g.player.ResetMoveCooldown()
	g.renderer.CenterCamera(g.player.X, g.player.Y)
	g.fov.Compute(dm, g.player.X, g.player.Y, fovRadius)
}

func (g *Game) playerAttack(enemy *Enemy) {
	result := ResolveAttack(g.player.Entity, enemy.Entity, g.rng)
	g.player.LastCombat = 0

	if result.IsDodge {
		g.particles.SpawnMiss(enemy.X, enemy.Y)
		return
	}
	if result.IsCrit {
		g.particles.SpawnCrit(enemy.X, enemy.Y, result.Damage)
	} else {
		g.particles.SpawnDamage(enemy.X, enemy.Y, result.Damage, "#ffffff")
	}

	if result.IsDeath {
		xp := XPForKill(enemy.BaseXP, g.floor)
		g.player.Stats.XP += xp
		g.killCount++
		g.particles.SpawnText(enemy.X, enemy.Y, "+"+itoa(xp)+"xp", "#FFD700")
	}
}

func (g *Game) enemyAttackPlayer(enemy *Enemy) {
	result := ResolveAttack(enemy.Entity, g.player.Entity, g.rng)
	g.player.LastCombat = 0

	if result.IsDodge {
		g.particles.SpawnMiss(g.player.X, g.player.Y)
		return
	}
	if result.IsCrit {
		g.particles.SpawnCrit(g.player.X, g.player.Y, result.Damage)
	} else {
		g.particles.SpawnDamage(g.player.X, g.player.Y, result.Damage, "#ff4444")
	}

	if result.IsDeath {
		g.state = StateDead
	}
}

func (g *Game) processEnemyCombat() {
	for _, e := range g.world.Enemies {
		if !e.IsAlive {
			continue
		}
		// Enemy attacks if adjacent to player
		dx := e.X - g.player.X
		dy := e.Y - g.player.Y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx <= 1 && dy <= 1 && (dx+dy) == 1 && e.ActionTimer <= 0 {
			g.enemyAttackPlayer(e)
			e.ActionTimer = e.Stats.MoveCooldownMS() / 1000.0
		}
	}
}

func (g *Game) processRegen(dt float64) {
	if g.player.LastCombat < regenDelay {
		return
	}
	g.player.RegenTimer += dt
	if g.player.RegenTimer >= regenInterval {
		g.player.RegenTimer -= regenInterval
		healed := g.player.Heal(1)
		if healed > 0 {
			g.particles.SpawnText(g.player.X, g.player.Y, "+1", "#44ff44")
		}
	}
}

func (g *Game) Update(dt float64) {
	switch g.state {
	case StatePlaying:
		g.player.Update(dt)
		g.processMovement()
		g.world.UpdateEnemies(dt, g.player.X, g.player.Y)
		g.processEnemyCombat()
		g.world.RemoveDead()
		g.processRegen(dt)
		g.particles.Update(dt)
	}
}

func (g *Game) Render() {
	g.renderer.Clear()

	switch g.state {
	case StateMenu:
		g.renderer.DrawText(30, 15, "G O R L", "#ffffff")
		g.renderer.DrawText(25, 18, "Go Roguelike", "#888888")
		g.renderer.DrawText(22, 25, "Press ENTER to start", "#aaaaaa")
	case StatePlaying:
		g.renderer.DrawDungeon(g.dungeonResult.Map, g.fov)

		// Draw enemies (only if visible)
		for _, e := range g.world.Enemies {
			if !e.IsAlive || !g.fov.IsVisible(e.X, e.Y) {
				continue
			}
			sprite := render.Sprites[e.Sprite]
			if sprite != nil {
				g.renderer.DrawSprite(sprite, 0, e.X, e.Y, sprite.Color)
			}
		}

		// Draw player sprite
		sprite := render.Sprites[g.player.Sprite]
		if sprite != nil {
			g.renderer.DrawSprite(sprite, 0, g.player.X, g.player.Y, sprite.Color)
		}

		// Draw particles
		g.particles.Draw(g.renderer)
	case StateDead:
		g.renderer.DrawText(28, 15, "YOU DIED", "#ff0000")
		g.renderer.DrawText(22, 18, "Floor: "+itoa(g.floor)+"  Kills: "+itoa(g.killCount), "#aaaaaa")
		g.renderer.DrawText(20, 25, "Press ENTER to restart", "#aaaaaa")
	}
}

// Run starts the game loop using requestAnimationFrame.
func (g *Game) Run() {
	var frame js.Func
	frame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		now := args[0].Float()
		if g.lastTime == 0 {
			g.lastTime = now
		}
		dt := (now - g.lastTime) / 1000.0
		g.lastTime = now

		g.Update(dt)
		g.Render()

		js.Global().Call("requestAnimationFrame", frame)
		return nil
	})
	js.Global().Call("requestAnimationFrame", frame)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	// reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
