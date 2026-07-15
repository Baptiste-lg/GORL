package dungeon

// FOV computes field of vision using symmetric recursive shadowcasting.
type FOV struct {
	Width, Height int
	Visible       [][]bool
	Explored      [][]bool
}

// NewFOV creates a new FOV tracker for the given map dimensions.
func NewFOV(w, h int) *FOV {
	visible := make([][]bool, h)
	explored := make([][]bool, h)
	for y := 0; y < h; y++ {
		visible[y] = make([]bool, w)
		explored[y] = make([]bool, w)
	}
	return &FOV{Width: w, Height: h, Visible: visible, Explored: explored}
}

// Compute calculates visible tiles from (ox, oy) with the given radius.
func (f *FOV) Compute(dm *DungeonMap, ox, oy, radius int) {
	// Reset visibility
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			f.Visible[y][x] = false
		}
	}

	// Origin is always visible
	f.setVisible(ox, oy)

	// Cast light in all 8 octants
	for oct := 0; oct < 8; oct++ {
		f.castLight(dm, ox, oy, radius, 1, 1.0, 0.0, oct)
	}
}

func (f *FOV) setVisible(x, y int) {
	if x >= 0 && x < f.Width && y >= 0 && y < f.Height {
		f.Visible[y][x] = true
		f.Explored[y][x] = true
	}
}

// IsVisible returns true if the tile at (x, y) is currently visible.
func (f *FOV) IsVisible(x, y int) bool {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return false
	}
	return f.Visible[y][x]
}

// IsExplored returns true if the tile at (x, y) has been seen before.
func (f *FOV) IsExplored(x, y int) bool {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return false
	}
	return f.Explored[y][x]
}

// ResetExplored clears all explored state (for new floors).
func (f *FOV) ResetExplored() {
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			f.Explored[y][x] = false
		}
	}
}

// octant multipliers for transforming coordinates
var octantDX = [8][2]int{
	{1, 0}, {0, 1}, {0, -1}, {1, 0},
	{-1, 0}, {0, -1}, {0, 1}, {-1, 0},
}
var octantDY = [8][2]int{
	{0, 1}, {1, 0}, {1, 0}, {0, -1},
	{0, -1}, {-1, 0}, {-1, 0}, {0, 1},
}

func (f *FOV) castLight(dm *DungeonMap, ox, oy, radius, row int, startSlope, endSlope float64, oct int) {
	if startSlope < endSlope {
		return
	}

	nextStart := startSlope
	for j := row; j <= radius; j++ {
		blocked := false
		for dx := -j; dx <= 0; dx++ {
			// Map column and row slopes
			lSlope := (float64(dx) - 0.5) / (float64(j) + 0.5)
			rSlope := (float64(dx) + 0.5) / (float64(j) - 0.5)

			if startSlope < rSlope {
				continue
			}
			if endSlope > lSlope {
				break
			}

			// Transform to map coordinates
			mx := ox + dx*octantDX[oct][0] + j*octantDX[oct][1]
			my := oy + dx*octantDY[oct][0] + j*octantDY[oct][1]

			// Check radius
			ddx := mx - ox
			ddy := my - oy
			if ddx*ddx+ddy*ddy > radius*radius {
				continue
			}

			f.setVisible(mx, my)

			if blocked {
				if dm.At(mx, my).BlocksSight() {
					nextStart = rSlope
					continue
				} else {
					blocked = false
					startSlope = nextStart
				}
			} else if dm.At(mx, my).BlocksSight() && j < radius {
				blocked = true
				f.castLight(dm, ox, oy, radius, j+1, startSlope, rSlope, oct)
				nextStart = rSlope
			}
		}
		if blocked {
			break
		}
	}
}
