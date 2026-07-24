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

	// Smooth camera (lerped pixel offset)
	smoothX, smoothY float64
	targetX, targetY float64
	cameraSmooth     float64

	// Screen shake
	shakeAmount   float64
	shakeDuration float64
	shakeTimer    float64
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
		ctx:          ctx,
		canvas:       canvas,
		cameraSmooth: 8.0,
	}
}

// CenterCamera sets the camera so that (tx, ty) is in the center of the viewport.
func (r *Renderer) CenterCamera(tx, ty int) {
	r.CamX = tx - ViewTilesX/2
	r.CamY = ty - ViewTilesY/2
	r.targetX = float64(r.CamX * TileCells * CellWidth)
	r.targetY = float64(r.CamY * TileCells * CellHeight)
}

// Shake triggers a screen shake effect.
func (r *Renderer) Shake(amount, duration float64) {
	r.shakeAmount = amount
	r.shakeDuration = duration
	r.shakeTimer = duration
}

// UpdateCamera updates smooth camera and shake. Call once per frame.
func (r *Renderer) UpdateCamera(dt float64) {
	// Lerp smooth camera
	r.smoothX += (r.targetX - r.smoothX) * r.cameraSmooth * dt
	r.smoothY += (r.targetY - r.smoothY) * r.cameraSmooth * dt

	// Update shake
	if r.shakeTimer > 0 {
		r.shakeTimer -= dt
		if r.shakeTimer < 0 {
			r.shakeTimer = 0
		}
	}
}

// Clear fills the entire canvas with black and applies shake offset.
func (r *Renderer) Clear() {
	r.ctx.Call("setTransform", 1, 0, 0, 1, 0, 0) // reset transform
	r.ctx.Set("fillStyle", "#000000")
	r.ctx.Call("fillRect", 0, 0, CanvasW, CanvasH)

	// Apply shake as canvas translate
	if r.shakeTimer > 0 {
		progress := r.shakeTimer / r.shakeDuration
		magnitude := r.shakeAmount * progress
		// Simple pseudo-random shake using timer value
		sx := magnitude * sinApprox(r.shakeTimer*73.0)
		sy := magnitude * sinApprox(r.shakeTimer*97.0)
		r.ctx.Call("setTransform", 1, 0, 0, 1, sx, sy)
	}
}

func sinApprox(x float64) float64 {
	// Fast sin approximation for shake
	for x > 6.28 {
		x -= 6.28
	}
	for x < -6.28 {
		x += 6.28
	}
	if x < 0 {
		return x * (1.27 + 0.405*x)
	}
	return x * (1.27 - 0.405*x)
}

// FillTileBg fills the 3x3 cell block for a tile with a background color.
func (r *Renderer) FillTileBg(vx, vy int, color string) {
	baseCol := vx * TileCells
	baseRow := vy * TileCells
	for dy := 0; dy < TileCells; dy++ {
		for dx := 0; dx < TileCells; dx++ {
			r.FillCell(baseCol+dx, baseRow+dy, color)
		}
	}
}

// DrawDungeonThemed renders the dungeon with a specific color theme.
// playerX/playerY are world coords for distance-based brightness.
func (r *Renderer) DrawDungeonThemed(dm *dungeon.DungeonMap, fov *dungeon.FOV, theme dungeon.Theme, playerX, playerY int) {
	for vy := 0; vy < ViewTilesY; vy++ {
		for vx := 0; vx < ViewTilesX; vx++ {
			wx := r.CamX + vx
			wy := r.CamY + vy

			tile := dm.At(wx, wy)
			if tile == dungeon.TileVoid {
				continue
			}

			col := vx*TileCells + 1
			row := vy*TileCells + 1

			if fov.IsVisible(wx, wy) {
				// Distance-based brightness falloff
				dx := wx - playerX
				dy := wy - playerY
				dist := dx*dx + dy*dy // squared distance
				brightness := 1.0
				if dist > 4 {
					brightness = 1.0 - float64(dist-4)*0.012
					if brightness < 0.35 {
						brightness = 0.35
					}
				}

				fg := tile.ThemedColor(theme)
				bg := tile.BgColor(theme)
				if brightness < 0.95 {
					fg = scaleColor(fg, brightness)
					bg = scaleColor(bg, brightness)
				}
				r.FillTileBg(vx, vy, bg)

				// Stairs get a glow effect
				if tile == dungeon.TileStairsDown || tile == dungeon.TileStairsUp {
					r.DrawCharGlow(col, row, tile.Glyph(), fg, "#ffdd44", 8)
				} else {
					r.DrawChar(col, row, tile.GlyphVariant(wx, wy), fg)
				}
			} else if fov.IsExplored(wx, wy) {
				r.FillTileBg(vx, vy, tile.DimBgColor(theme))
				r.DrawChar(col, row, tile.GlyphVariant(wx, wy), dimColor(tile.ThemedColor(theme)))
			}
		}
	}
}

// scaleColor multiplies each RGB channel by a brightness factor (0.0-1.0).
func scaleColor(hex string, brightness float64) string {
	if len(hex) != 7 {
		return hex
	}
	r := hexVal(hex[1])<<4 + hexVal(hex[2])
	g := hexVal(hex[3])<<4 + hexVal(hex[4])
	b := hexVal(hex[5])<<4 + hexVal(hex[6])
	r = int(float64(r) * brightness)
	g = int(float64(g) * brightness)
	b = int(float64(b) * brightness)
	return "#" + hexByte(r) + hexByte(g) + hexByte(b)
}

// dimColor returns a darker version of a hex color for explored-but-not-visible tiles.
func dimColor(hex string) string {
	if len(hex) != 7 {
		return "#222222"
	}
	r := hexVal(hex[1])<<4 + hexVal(hex[2])
	g := hexVal(hex[3])<<4 + hexVal(hex[4])
	b := hexVal(hex[5])<<4 + hexVal(hex[6])
	// Dim to 40% — visible but clearly darker than lit areas
	r = r * 2 / 5
	g = g * 2 / 5
	b = b * 2 / 5
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

// DrawCharGlow renders a character with a glow halo effect.
func (r *Renderer) DrawCharGlow(col, row int, ch string, color string, glowColor string, blur int) {
	px := col * CellWidth
	py := row * CellHeight
	r.ctx.Set("shadowColor", glowColor)
	r.ctx.Set("shadowBlur", blur)
	r.ctx.Set("fillStyle", color)
	r.ctx.Call("fillText", ch, px, py)
	r.ctx.Set("shadowBlur", 0)
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
func (r *Renderer) DrawSprite(sprite *Sprite, frame int, worldX, worldY int, color string) {
	r.DrawSpriteWithBg(sprite, frame, worldX, worldY, color, "")
}

// FlashTile briefly highlights a tile with a bright color (for hit effects).
func (r *Renderer) FlashTile(worldX, worldY int, flashColor string) {
	vx := worldX - r.CamX
	vy := worldY - r.CamY
	if vx < 0 || vx >= ViewTilesX || vy < 0 || vy >= ViewTilesY {
		return
	}
	r.FillTileBg(vx, vy, flashColor)
}

// DrawSpriteWithBg renders a sprite with an optional background tint behind it.
func (r *Renderer) DrawSpriteWithBg(sprite *Sprite, frame int, worldX, worldY int, color, bgColor string) {
	vx := worldX - r.CamX
	vy := worldY - r.CamY

	if vx < 0 || vx >= ViewTilesX || vy < 0 || vy >= ViewTilesY {
		return
	}

	baseCol := vx * TileCells
	baseRow := vy * TileCells

	// Draw background tint if specified
	if bgColor != "" {
		for dy := 0; dy < TileCells; dy++ {
			for dx := 0; dx < TileCells; dx++ {
				r.FillCell(baseCol+dx, baseRow+dy, bgColor)
			}
		}
	}

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
