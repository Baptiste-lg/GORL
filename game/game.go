package game

import (
	"syscall/js"

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
}

func New() *Game {
	g := &Game{
		state:    StateMenu,
		renderer: render.NewRenderer(),
		keys:     make(map[string]bool),
	}
	g.registerInput()
	return g
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
			g.state = StatePlaying
		}
	case StateDead:
		if key == "Enter" {
			g.state = StateMenu
		}
	}
}

func (g *Game) Update(dt float64) {
	switch g.state {
	case StatePlaying:
		// Game logic will go here
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
		g.renderer.DrawText(35, 20, "@", "#00ff00")
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
		dt := (now - g.lastTime) / 1000.0 // seconds
		g.lastTime = now

		g.Update(dt)
		g.Render()

		js.Global().Call("requestAnimationFrame", frame)
		return nil
	})
	js.Global().Call("requestAnimationFrame", frame)
}
