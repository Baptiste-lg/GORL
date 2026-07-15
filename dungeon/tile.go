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

// Color returns the display color for this tile.
func (t Tile) Color() string {
	switch t {
	case TileWall:
		return "#666666"
	case TileFloor:
		return "#333333"
	case TileDoor:
		return "#8B4513"
	case TileStairsDown, TileStairsUp:
		return "#FFD700"
	default:
		return "#000000"
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
