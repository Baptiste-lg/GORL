package render

import "syscall/js"

const (
	CellWidth  = 16
	CellHeight = 20
	GridCols   = 80
	GridRows   = 40
	CanvasW    = GridCols * CellWidth  // 1280
	CanvasH    = GridRows * CellHeight // 800
	FontSize   = 16
	FontFace   = "16px monospace"
)

type Renderer struct {
	ctx    js.Value
	canvas js.Value
	// Camera offset in tile coordinates
	CamX int
	CamY int
}

func NewRenderer() *Renderer {
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "game-canvas")
	ctx := canvas.Call("getContext", "2d")

	canvas.Set("width", CanvasW)
	canvas.Set("height", CanvasH)

	ctx.Set("font", FontFace)
	ctx.Set("textBaseline", "top")

	return &Renderer{
		ctx:    ctx,
		canvas: canvas,
	}
}

// Clear fills the entire canvas with black.
func (r *Renderer) Clear() {
	r.ctx.Set("fillStyle", "#000000")
	r.ctx.Call("fillRect", 0, 0, CanvasW, CanvasH)
}

// DrawChar renders a single character at grid position (col, row) with the given color.
func (r *Renderer) DrawChar(col, row int, ch string, color string) {
	px := col * CellWidth
	py := row * CellHeight
	r.ctx.Set("fillStyle", color)
	r.ctx.Call("fillText", ch, px, py)
}

// DrawText renders a string starting at grid position (col, row).
func (r *Renderer) DrawText(col, row int, text string, color string) {
	px := col * CellWidth
	py := row * CellHeight
	r.ctx.Set("fillStyle", color)
	r.ctx.Call("fillText", text, px, py)
}

// FillRect fills a rectangle in pixel coordinates.
func (r *Renderer) FillRect(x, y, w, h int, color string) {
	r.ctx.Set("fillStyle", color)
	r.ctx.Call("fillRect", x, y, w, h)
}

// FillCell fills a single cell at grid position (col, row).
func (r *Renderer) FillCell(col, row int, color string) {
	r.FillRect(col*CellWidth, row*CellHeight, CellWidth, CellHeight, color)
}
