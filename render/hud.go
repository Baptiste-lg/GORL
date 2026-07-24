package render

// HUDData holds everything the HUD needs to display.
type HUDData struct {
	HP, MaxHP   int
	Level       int
	XP, XPNext  int
	Floor       int
	Gold        int
	ActiveName  string // active item name (empty if none)
	ActiveReady bool   // true if fully charged
	ActivePct   int    // charge percentage 0-100
	Effects     []HUDEffect
	Streak      string // kill streak label (empty if none)
}

// HUDEffect is a status effect shown in the HUD.
type HUDEffect struct {
	Name      string
	Remaining float64
}

// barBuf is reused to build bar strings without allocating per frame.
var barBuf [22]byte // "[" + 20 chars + "]"

func buildBar(filled, width int, fillChar, emptyChar byte) string {
	barBuf[0] = '['
	for i := 0; i < width; i++ {
		if i < filled {
			barBuf[1+i] = fillChar
		} else {
			barBuf[1+i] = emptyChar
		}
	}
	barBuf[1+width] = ']'
	return string(barBuf[:2+width])
}

// DrawHUD renders the heads-up display overlay.
func (r *Renderer) DrawHUD(d HUDData) {
	// === HUD background panels ===
	// Top bar background (rows 0-4)
	for row := 0; row < 5; row++ {
		for col := 0; col < GridCols; col++ {
			r.FillCell(col, row, "#0a0a0e")
		}
	}
	// Bottom bar background (last row)
	if len(d.Effects) > 0 {
		for col := 0; col < GridCols; col++ {
			r.FillCell(col, GridRows-1, "#0a0a0e")
		}
	}

	// === Top-left: HP bar ===
	r.DrawText(1, 0, "HP", "#ff4444")
	barW := 20
	filled := 0
	if d.MaxHP > 0 {
		filled = d.HP * barW / d.MaxHP
	}
	r.DrawText(4, 0, buildBar(filled, barW, '|', ' '), "#ff4444")
	r.DrawText(26, 0, intToStr(d.HP)+"/"+intToStr(d.MaxHP), "#ff8888")

	// === Top-left row 2: XP bar ===
	r.DrawText(1, 1, "XP", "#FFD700")
	xpFilled := 0
	if d.XPNext > 0 {
		xpFilled = d.XP * barW / d.XPNext
		if xpFilled > barW {
			xpFilled = barW
		}
	}
	r.DrawText(4, 1, buildBar(xpFilled, barW, ':', ' '), "#FFD700")
	r.DrawText(26, 1, intToStr(d.XP)+"/"+intToStr(d.XPNext), "#ccaa00")

	// === Top-right: Floor, Level, Gold ===
	floorStr := "Floor " + intToStr(d.Floor)
	levelStr := "Lv." + intToStr(d.Level)
	goldStr := "$" + intToStr(d.Gold)
	r.DrawText(GridCols-len(floorStr)-1, 0, floorStr, "#aaaaaa")
	r.DrawText(GridCols-len(levelStr)-1, 1, levelStr, "#ffffff")
	r.DrawText(GridCols-len(goldStr)-1, 2, goldStr, "#FFD700")

	// === Active item (row 3, right-aligned) ===
	if d.ActiveName != "" {
		label := "[Q] " + d.ActiveName + " "
		barStart := GridCols - len(label) - 12
		r.DrawText(barStart, 3, label, "#aaaaaa")
		// Visual charge bar (10 chars wide)
		chargeFilled := d.ActivePct / 10
		if d.ActiveReady {
			r.DrawText(barStart+len(label), 3, buildBar(10, 10, '=', '='), "#44ffaa")
			r.DrawText(barStart+len(label)+12, 3, "READY!", "#44ffaa")
		} else {
			chargeColor := "#886600"
			if d.ActivePct >= 80 {
				chargeColor = "#ccaa00"
			}
			r.DrawText(barStart+len(label), 3, buildBar(chargeFilled, 10, '|', ' '), chargeColor)
		}
	}

	// === Kill streak ===
	if d.Streak != "" {
		r.DrawText((GridCols-len(d.Streak))/2, 4, d.Streak, "#ff8800")
	}

	// === Bottom: status effects ===
	for i, eff := range d.Effects {
		col := 1 + i*20
		secs := int(eff.Remaining) + 1
		r.DrawText(col, GridRows-1, eff.Name+" "+intToStr(secs)+"s", "#44aaff")
	}
}

// MiniMapMarker is a colored dot on the minimap.
type MiniMapMarker struct {
	X, Y  int
	Color string
}

// MiniMapData holds what the mini-map needs.
type MiniMapData struct {
	MapW, MapH int
	PlayerX    int
	PlayerY    int
	IsExplored func(x, y int) bool
	IsWall     func(x, y int) bool
	Markers    []MiniMapMarker
}

// DrawMiniMap renders a small overview in the bottom-right corner.
func (r *Renderer) DrawMiniMap(d MiniMapData) {
	// Mini-map size in cells
	mmW := d.MapW / 2
	mmH := d.MapH / 2
	if mmW > 28 {
		mmW = 28
	}
	if mmH > 12 {
		mmH = 12
	}

	// Position: bottom-right corner
	ox := GridCols - mmW - 1
	oy := GridRows - mmH - 2

	// Background
	for dy := 0; dy < mmH; dy++ {
		for dx := 0; dx < mmW; dx++ {
			r.FillCell(ox+dx, oy+dy, "#0a0a0a")
		}
	}

	// Draw explored tiles
	scaleX := float64(d.MapW) / float64(mmW)
	scaleY := float64(d.MapH) / float64(mmH)

	for dy := 0; dy < mmH; dy++ {
		for dx := 0; dx < mmW; dx++ {
			wx := int(float64(dx) * scaleX)
			wy := int(float64(dy) * scaleY)

			if !d.IsExplored(wx, wy) {
				continue
			}

			color := "#222222"
			if d.IsWall(wx, wy) {
				color = "#444444"
			}
			r.FillCell(ox+dx, oy+dy, color)
		}
	}

	// Special room markers (only if explored)
	for _, m := range d.Markers {
		if !d.IsExplored(m.X, m.Y) {
			continue
		}
		mx := int(float64(m.X) / scaleX)
		my := int(float64(m.Y) / scaleY)
		if mx >= 0 && mx < mmW && my >= 0 && my < mmH {
			r.FillCell(ox+mx, oy+my, m.Color)
		}
	}

	// Player dot (drawn last, on top)
	px := int(float64(d.PlayerX) / scaleX)
	py := int(float64(d.PlayerY) / scaleY)
	if px >= 0 && px < mmW && py >= 0 && py < mmH {
		r.FillCell(ox+px, oy+py, "#00ff00")
	}
}
