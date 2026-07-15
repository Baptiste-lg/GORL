package render

// DrawBox draws an ASCII-bordered box at grid position (col, row) with given dimensions.
func (r *Renderer) DrawBox(col, row, w, h int, borderColor, bgColor string) {
	// Fill background
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			r.FillCell(col+dx, row+dy, bgColor)
		}
	}

	// Top and bottom borders
	r.DrawChar(col, row, "+", borderColor)
	r.DrawChar(col+w-1, row, "+", borderColor)
	r.DrawChar(col, row+h-1, "+", borderColor)
	r.DrawChar(col+w-1, row+h-1, "+", borderColor)
	for x := 1; x < w-1; x++ {
		r.DrawChar(col+x, row, "-", borderColor)
		r.DrawChar(col+x, row+h-1, "-", borderColor)
	}
	for y := 1; y < h-1; y++ {
		r.DrawChar(col, row+y, "|", borderColor)
		r.DrawChar(col+w-1, row+y, "|", borderColor)
	}
}

// InventoryData holds what the UI needs to render the inventory screen.
type InventoryData struct {
	Items       []InventoryItem
	WeaponName  string
	WeaponColor string
	ArmorName   string
	ArmorColor  string
	Selected    int
}

// InventoryItem is a simplified item for rendering.
type InventoryItem struct {
	Name   string
	Glyph  string
	Color  string
	IsSel  bool
	Detail string // e.g. "+3 STR +1 DEX"
}

// DrawInventory renders the inventory overlay.
func (r *Renderer) DrawInventory(data InventoryData) {
	boxW, boxH := 40, 28
	ox, oy := (GridCols-boxW)/2, (GridRows-boxH)/2

	r.DrawBox(ox, oy, boxW, boxH, "#888888", "#111111")

	// Title
	r.DrawText(ox+14, oy+1, "INVENTORY", "#ffffff")

	// Equipment slots
	r.DrawText(ox+2, oy+3, "Equipment:", "#aaaaaa")
	wepName := "(empty)"
	wepColor := "#555555"
	if data.WeaponName != "" {
		wepName = data.WeaponName
		wepColor = data.WeaponColor
	}
	r.DrawText(ox+3, oy+4, "Weapon: "+wepName, wepColor)

	armName := "(empty)"
	armColor := "#555555"
	if data.ArmorName != "" {
		armName = data.ArmorName
		armColor = data.ArmorColor
	}
	r.DrawText(ox+3, oy+5, "Armor:  "+armName, armColor)

	// Separator
	for x := 1; x < boxW-1; x++ {
		r.DrawChar(ox+x, oy+7, "-", "#444444")
	}

	// Items
	r.DrawText(ox+2, oy+8, "Backpack:", "#aaaaaa")
	if len(data.Items) == 0 {
		r.DrawText(ox+3, oy+9, "(empty)", "#555555")
	}
	for i, item := range data.Items {
		y := oy + 9 + i
		prefix := "  "
		if i == data.Selected {
			prefix = "> "
		}
		r.DrawText(ox+2, y, prefix+item.Glyph+" "+item.Name, item.Color)
	}

	// Tooltip for selected item
	if data.Selected >= 0 && data.Selected < len(data.Items) {
		item := data.Items[data.Selected]
		tipY := oy + 9 + len(data.Items) + 1
		if item.Detail != "" {
			r.DrawText(ox+3, tipY, item.Detail, "#aaaaaa")
		}
	}

	// Controls
	r.DrawText(ox+2, oy+boxH-3, "Arrow keys: navigate", "#555555")
	r.DrawText(ox+2, oy+boxH-2, "Enter: equip/use  D: drop  I: close", "#555555")
}

// LevelUpChoice holds one stat boost option.
type LevelUpChoice struct {
	Label string
	Desc  string
}

// DrawLevelUp renders the level-up screen with 3 choices.
func (r *Renderer) DrawLevelUp(level int, choices []LevelUpChoice, selected int) {
	boxW, boxH := 40, 16
	ox, oy := (GridCols-boxW)/2, (GridRows-boxH)/2

	r.DrawBox(ox, oy, boxW, boxH, "#FFD700", "#111111")

	r.DrawText(ox+12, oy+1, "LEVEL UP!", "#FFD700")
	r.DrawText(ox+14, oy+2, "Lv."+intToStr(level), "#ffffff")

	r.DrawText(ox+2, oy+4, "Choose a stat boost:", "#aaaaaa")

	for i, ch := range choices {
		y := oy + 6 + i*3
		prefix := "  "
		color := "#cccccc"
		if i == selected {
			prefix = "> "
			color = "#FFD700"
		}
		r.DrawText(ox+3, y, prefix+ch.Label, color)
		r.DrawText(ox+6, y+1, ch.Desc, "#888888")
	}

	r.DrawText(ox+2, oy+boxH-2, "Enter: confirm", "#555555")
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 5)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
