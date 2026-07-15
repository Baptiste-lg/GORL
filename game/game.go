package game

import (
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

const fovRadius = 8

type Game struct {
	state    State
	renderer *render.Renderer
	lastTime float64
	keys     map[string]bool

	dungeonResult *dungeon.GenerateResult
	fov           *dungeon.FOV
	player        *Player
	floor         int
}

func New() *Game {
	g := &Game{
		state:    StateMenu,
		renderer: render.NewRenderer(),
		keys:     make(map[string]bool),
		floor:    1,
	}
	g.registerInput()
	return g
}

func (g *Game) startNewGame() {
	g.floor = 1
	g.generateFloor()
	g.state = StatePlaying
}

func (g *Game) generateFloor() {
	seed := time.Now().UnixNano()
	g.dungeonResult = dungeon.Generate(seed)
	g.fov = dungeon.NewFOV(dungeon.MapWidth, dungeon.MapHeight)

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
	if dm.At(nx, ny).Passable() {
		g.player.X = nx
		g.player.Y = ny
		g.player.ResetMoveCooldown()
		g.renderer.CenterCamera(g.player.X, g.player.Y)
		g.fov.Compute(dm, g.player.X, g.player.Y, fovRadius)
	}
}

func (g *Game) Update(dt float64) {
	switch g.state {
	case StatePlaying:
		g.player.Update(dt)
		g.processMovement()
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
		// Draw player sprite
		sprite := render.Sprites[g.player.Sprite]
		if sprite != nil {
			g.renderer.DrawSprite(sprite, 0, g.player.X, g.player.Y, sprite.Color)
		}
	case StateDead:
		g.renderer.DrawText(28, 15, "YOU DIED", "#ff0000")
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
