package dungeon

type Tile int

const (
	TileVoid Tile = iota
	TileWall
	TileFloor
	TileDoor
	TileStairsDown
	TileStairsUp
)

// Glyph returns the ASCII character for this tile.
func (t Tile) Glyph() string {
	switch t {
	case TileWall:
		return "#"
	case TileFloor:
		return "."
	case TileDoor:
		return "+"
	case TileStairsDown:
		return ">"
	case TileStairsUp:
		return "<"
	default:
		return " "
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

// themeColors maps [theme][tile] to a color string.
var themeColors = [4][5]string{
	// Stone: wall, floor, door, stairs, void
	{"#666666", "#333333", "#8B4513", "#FFD700", "#000000"},
	// Crypt
	{"#446644", "#223322", "#556B2F", "#FFD700", "#000000"},
	// Inferno
	{"#884422", "#331111", "#AA4400", "#FFD700", "#000000"},
	// Abyss
	{"#443366", "#1a0a2e", "#6633AA", "#FFD700", "#000000"},
}

// Color returns the display color for this tile.
func (t Tile) Color() string {
	return t.ThemedColor(ThemeStone)
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
	return t == TileWall || t == TileVoid
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
