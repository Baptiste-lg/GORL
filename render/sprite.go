package render

// Sprite holds animation frames as multi-line ASCII art.
// Each frame is a slice of strings (one per row).
type Sprite struct {
	Frames [][]string
	Color  string
	Width  int // number of characters wide
	Height int // number of lines tall
}

// Frame returns the frame at index (wraps around).
func (s *Sprite) Frame(idx int) []string {
	return s.Frames[idx%len(s.Frames)]
}

// All sprites are defined here as multi-char ASCII art.
var Sprites = map[string]*Sprite{
	"player": {
		Frames: [][]string{
			{" O ", "/|\\", "/ \\"},  // idle
			{" O ", "/|\\", "| |"},  // walk
		},
		Color:  "#00ff00",
		Width:  3,
		Height: 3,
	},
	"rat": {
		Frames: [][]string{
			{"   ", "o--", " ''"},
		},
		Color:  "#8B4513",
		Width:  3,
		Height: 3,
	},
	"skeleton": {
		Frames: [][]string{
			{"._.", "/X\\", "| |"},
			{"._.", "/X\\", "/ \\"},
		},
		Color:  "#cccccc",
		Width:  3,
		Height: 3,
	},
	"bat": {
		Frames: [][]string{
			{"/V\\", "\\_/", "   "},
			{"/v\\", "/ \\", "   "},
		},
		Color:  "#9966cc",
		Width:  3,
		Height: 3,
	},
	"slime": {
		Frames: [][]string{
			{"___", "(o_o)", "~~~"},
		},
		Color:  "#33cc33",
		Width:  3,
		Height: 3,
	},
	"ghost": {
		Frames: [][]string{
			{"___", "(O.O)", "^^^"},
			{"___", "(o.o)", "^^^"},
		},
		Color:  "#aaaaff",
		Width:  3,
		Height: 3,
	},
	"chest": {
		Frames: [][]string{
			{"[===]", "|___|", "     "},
		},
		Color:  "#FFD700",
		Width:  5,
		Height: 3,
	},
}
