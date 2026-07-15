package render

// HUDData holds everything the HUD needs to display.
type HUDData struct {
	HP, MaxHP     int
	Level         int
	XP, XPNext    int
	Floor         int
	Effects       []HUDEffect
}

// HUDEffect is a status effect shown in the HUD.
type HUDEffect struct {
	Name      string
	Remaining float64
}

// DrawHUD renders the heads-up display overlay.
func (r *Renderer) DrawHUD(d HUDData) {
	// === Top-left: HP bar ===
	r.DrawText(1, 0, "HP", "#ff4444")
	barW := 20
	filled := 0
	if d.MaxHP > 0 {
		filled = d.HP * barW / d.MaxHP
	}
	bar := "["
	for i := 0; i < barW; i++ {
		if i < filled {
			bar += "|"
		} else {
			bar += " "
		}
	}
	bar += "]"
	r.DrawText(4, 0, bar, "#ff4444")
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
	xpBar := "["
	for i := 0; i < barW; i++ {
		if i < xpFilled {
			xpBar += ":"
		} else {
			xpBar += " "
		}
	}
	xpBar += "]"
	r.DrawText(4, 1, xpBar, "#FFD700")
	r.DrawText(26, 1, intToStr(d.XP)+"/"+intToStr(d.XPNext), "#ccaa00")

	// === Top-right: Floor and Level ===
	floorStr := "Floor " + intToStr(d.Floor)
	levelStr := "Lv." + intToStr(d.Level)
	r.DrawText(GridCols-len(floorStr)-1, 0, floorStr, "#aaaaaa")
	r.DrawText(GridCols-len(levelStr)-1, 1, levelStr, "#ffffff")

	// === Bottom: status effects ===
	for i, eff := range d.Effects {
		col := 1 + i*20
		secs := int(eff.Remaining) + 1
		r.DrawText(col, GridRows-1, eff.Name+" "+intToStr(secs)+"s", "#44aaff")
	}
}
