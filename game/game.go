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

type Game struct {
	state    State
	renderer *render.Renderer
	lastTime float64
	keys     map[string]bool

	dungeonResult *dungeon.GenerateResult
	fov           *dungeon.FOV
	playerX       int
	playerY       int
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

const fovRadius = 8

func (g *Game) generateFloor() {
	seed := time.Now().UnixNano()
	g.dungeonResult = dungeon.Generate(seed)
	g.fov = dungeon.NewFOV(dungeon.MapWidth, dungeon.MapHeight)
	g.playerX = g.dungeonResult.SpawnX
	g.playerY = g.dungeonResult.SpawnY
	g.renderer.CenterCamera(g.playerX, g.playerY)
	g.fov.Compute(g.dungeonResult.Map, g.playerX, g.playerY, fovRadius)
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
	case StatePlaying:
		dx, dy := 0, 0
		switch key {
		case "w", "ArrowUp":
			dy = -1
		case "s", "ArrowDown":
			dy = 1
		case "a", "ArrowLeft":
			dx = -1
		case "d", "ArrowRight":
			dx = 1
		}
		if dx != 0 || dy != 0 {
			g.tryMove(dx, dy)
		}
	}
}

func (g *Game) tryMove(dx, dy int) {
	nx, ny := g.playerX+dx, g.playerY+dy
	dm := g.dungeonResult.Map
	if dm.At(nx, ny).Passable() {
		g.playerX = nx
		g.playerY = ny
		g.renderer.CenterCamera(g.playerX, g.playerY)
		g.fov.Compute(g.dungeonResult.Map, g.playerX, g.playerY, fovRadius)
	}
}

func (g *Game) Update(dt float64) {
	switch g.state {
	case StatePlaying:
		// Real-time game logic will go here
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
		// Draw player at center of viewport
		vcx := render.ViewTilesX / 2
		vcy := render.ViewTilesY / 2
		col := vcx*render.TileCells + 1
		row := vcy*render.TileCells + 1
		g.renderer.DrawChar(col, row, "@", "#00ff00")
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
