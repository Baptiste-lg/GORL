package render

import (
	"syscall/js"

	"github.com/Baptiste-lg/GORL/dungeon"
)

const (
	CellWidth  = 16
	CellHeight = 20
	GridCols   = 80
	GridRows   = 40
	CanvasW    = GridCols * CellWidth  // 1280
	CanvasH    = GridRows * CellHeight // 800
	FontSize   = 16
	FontFace   = "16px monospace"

	// Each world tile occupies a 3x3 cell block for multi-char sprites.
	TileCells = 3
	// Visible tiles in the viewport.
	ViewTilesX = GridCols / TileCells // ~26
	ViewTilesY = GridRows / TileCells // ~13
)

type Renderer struct {
	ctx    js.Value
	canvas js.Value
	// Camera center in tile coordinates
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

// CenterCamera sets the camera so that (tx, ty) is in the center of the viewport.
func (r *Renderer) CenterCamera(tx, ty int) {
	r.CamX = tx - ViewTilesX/2
	r.CamY = ty - ViewTilesY/2
}

// Clear fills the entire canvas with black.
func (r *Renderer) Clear() {
	r.ctx.Set("fillStyle", "#000000")
	r.ctx.Call("fillRect", 0, 0, CanvasW, CanvasH)
}

// DrawDungeon renders the visible portion of the dungeon map with FOV.
func (r *Renderer) DrawDungeon(dm *dungeon.DungeonMap, fov *dungeon.FOV) {
	for vy := 0; vy < ViewTilesY; vy++ {
		for vx := 0; vx < ViewTilesX; vx++ {
			wx := r.CamX + vx
			wy := r.CamY + vy

			tile := dm.At(wx, wy)
			if tile == dungeon.TileVoid {
				continue
			}

			// Screen cell position (center of the 3x3 block)
			col := vx*TileCells + 1
			row := vy*TileCells + 1

			if fov.IsVisible(wx, wy) {
				r.DrawChar(col, row, tile.Glyph(), tile.Color())
			} else if fov.IsExplored(wx, wy) {
				r.DrawChar(col, row, tile.Glyph(), dimColor(tile.Color()))
			}
		}
	}
}

// dimColor returns a darker version of a hex color for explored-but-not-visible tiles.
func dimColor(hex string) string {
	if len(hex) != 7 {
		return "#111111"
	}
	// Simple approach: halve each component and halve again
	r := hexVal(hex[1])<<4 + hexVal(hex[2])
	g := hexVal(hex[3])<<4 + hexVal(hex[4])
	b := hexVal(hex[5])<<4 + hexVal(hex[6])
	r /= 3
	g /= 3
	b /= 3
	return "#" + hexByte(r) + hexByte(g) + hexByte(b)
}

func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return 0
}

func hexByte(v int) string {
	const digits = "0123456789abcdef"
	if v > 255 {
		v = 255
	}
	return string([]byte{digits[v>>4], digits[v&0xf]})
}

// DrawChar renders a single character at grid position (col, row).
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

// DrawSprite renders a multi-char sprite at a world tile position.
// The sprite is drawn within the 3x3 cell block for that tile.
// worldX/worldY are tile coords; the renderer converts to screen coords using the camera.
func (r *Renderer) DrawSprite(sprite *Sprite, frame int, worldX, worldY int, color string) {
	// Convert world tile to viewport tile
	vx := worldX - r.CamX
	vy := worldY - r.CamY

	if vx < 0 || vx >= ViewTilesX || vy < 0 || vy >= ViewTilesY {
		return // off screen
	}

	// Top-left cell of the 3x3 block
	baseCol := vx * TileCells
	baseRow := vy * TileCells

	lines := sprite.Frame(frame)
	for row, line := range lines {
		for col, ch := range line {
			if ch == ' ' {
				continue
			}
			r.DrawChar(baseCol+col, baseRow+row, string(ch), color)
		}
	}
}

// DrawSpriteAt renders a sprite at an absolute screen cell position (not world-relative).
func (r *Renderer) DrawSpriteAt(sprite *Sprite, frame int, col, row int, color string) {
	lines := sprite.Frame(frame)
	for dy, line := range lines {
		for dx, ch := range line {
			if ch == ' ' {
				continue
			}
			r.DrawChar(col+dx, row+dy, string(ch), color)
		}
	}
}
