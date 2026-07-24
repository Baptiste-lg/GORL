package dungeon

type Tile int

const (
	TileVoid Tile = iota
	TileWall
	TileFloor
	TileDoor
	TileStairsDown
	TileStairsUp
	TileCrackedWall
)

// Glyph returns the display character for this tile.
func (t Tile) Glyph() string {
	switch t {
	case TileWall:
		return "▓"
	case TileCrackedWall:
		return "▒"
	case TileFloor:
		return "·"
	case TileDoor:
		return "▌"
	case TileStairsDown:
		return ">"
	case TileStairsUp:
		return "<"
	default:
		return " "
	}
}

// GlyphVariant returns a position-seeded glyph variant for visual variety.
func (t Tile) GlyphVariant(x, y int) string {
	switch t {
	case TileFloor:
		h := (x*7 + y*13) & 15
		if h == 0 {
			return ","
		}
		if h == 1 {
			return "∙"
		}
		if h == 2 {
			return " " // some floor tiles are blank (just bg)
		}
		return "·"
	case TileWall:
		h := (x*11 + y*17) & 7
		if h == 0 {
			return "█"
		}
		if h == 1 {
			return "▓"
		}
		return "▓"
	case TileCrackedWall:
		h := (x*5 + y*9) & 3
		if h == 0 {
			return "░"
		}
		return "▒"
	default:
		return t.Glyph()
	}
}

// Theme controls dungeon color palette.
type Theme int

const (
	ThemeStone   Theme = iota // floors 1-4
	ThemeCrypt                // floors 5-9
	ThemeInferno              // floors 10-14
	ThemeAbyss                // floors 15+
)

// ThemeForFloor returns the appropriate theme.
func ThemeForFloor(floor int) Theme {
	switch {
	case floor >= 15:
		return ThemeAbyss
	case floor >= 10:
		return ThemeInferno
	case floor >= 5:
		return ThemeCrypt
	default:
		return ThemeStone
	}
}

// themeColors maps [theme][tile] to a foreground color string.
var themeColors = [4][5]string{
	// Stone: wall, floor, door, stairs, void
	{"#8a8a7a", "#6a6a60", "#b08040", "#ffee55", "#000000"},
	// Crypt
	{"#5a8a5a", "#4a7a4a", "#88a050", "#ffee55", "#000000"},
	// Inferno
	{"#aa6644", "#886655", "#dd7722", "#ffee55", "#000000"},
	// Abyss
	{"#7766aa", "#665599", "#9966dd", "#ffee55", "#000000"},
}

// themeBgColors maps [theme][tile] to a background color string.
// wall, floor, door, stairs, explored-not-visible
var themeBgColors = [4][5]string{
	// Stone — strong wall/floor contrast
	{"#2a2a32", "#141418", "#241c12", "#1a1a20", "#0c0c10"},
	// Crypt
	{"#1e2e1e", "#0c160c", "#1e2412", "#0c160c", "#081008"},
	// Inferno
	{"#301c12", "#1a0e0a", "#28180c", "#1a0e0a", "#100a06"},
	// Abyss
	{"#241c44", "#120e22", "#1e1836", "#120e22", "#0a081a"},
}

// BgColor returns the background color for this tile in the given theme.
func (t Tile) BgColor(theme Theme) string {
	idx := int(theme)
	if idx < 0 || idx > 3 {
		idx = 0
	}
	switch t {
	case TileWall, TileCrackedWall:
		return themeBgColors[idx][0]
	case TileFloor:
		return themeBgColors[idx][1]
	case TileDoor:
		return themeBgColors[idx][2]
	case TileStairsDown, TileStairsUp:
		return themeBgColors[idx][3]
	default:
		return "#000000"
	}
}

// DimBgColor returns a very dark background for explored-but-not-visible tiles.
func (t Tile) DimBgColor(theme Theme) string {
	idx := int(theme)
	if idx < 0 || idx > 3 {
		idx = 0
	}
	if t == TileVoid {
		return "#000000"
	}
	return themeBgColors[idx][4]
}

// ThemedColor returns the display color for this tile in a given theme.
func (t Tile) ThemedColor(theme Theme) string {
	idx := int(theme)
	if idx < 0 || idx > 3 {
		idx = 0
	}
	switch t {
	case TileWall:
		return themeColors[idx][0]
	case TileCrackedWall:
		return "#aa9966"
	case TileFloor:
		return themeColors[idx][1]
	case TileDoor:
		return themeColors[idx][2]
	case TileStairsDown, TileStairsUp:
		return themeColors[idx][3]
	default:
		return themeColors[idx][4]
	}
}

// Passable returns true if entities can walk on this tile.
func (t Tile) Passable() bool {
	return t == TileFloor || t == TileDoor || t == TileStairsDown || t == TileStairsUp
}

// BlocksSight returns true if this tile blocks line of sight.
func (t Tile) BlocksSight() bool {
	return t == TileWall || t == TileVoid || t == TileCrackedWall
}

// DungeonMap holds the 2D grid of tiles for a single floor.
type DungeonMap struct {
	Width  int
	Height int
	Tiles  [][]Tile
}

// NewDungeonMap creates a map filled with walls.
func NewDungeonMap(w, h int) *DungeonMap {
	tiles := make([][]Tile, h)
	for y := range tiles {
		tiles[y] = make([]Tile, w)
		for x := range tiles[y] {
			tiles[y][x] = TileWall
		}
	}
	return &DungeonMap{Width: w, Height: h, Tiles: tiles}
}

// InBounds checks if (x, y) is within the map.
func (d *DungeonMap) InBounds(x, y int) bool {
	return x >= 0 && x < d.Width && y >= 0 && y < d.Height
}

// At returns the tile at (x, y), or TileVoid if out of bounds.
func (d *DungeonMap) At(x, y int) Tile {
	if !d.InBounds(x, y) {
		return TileVoid
	}
	return d.Tiles[y][x]
}

// Set sets the tile at (x, y).
func (d *DungeonMap) Set(x, y int, t Tile) {
	if d.InBounds(x, y) {
		d.Tiles[y][x] = t
	}
}
