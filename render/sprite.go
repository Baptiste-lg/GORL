package render

// Sprite holds animation frames as multi-line ASCII art.
// Each frame is a slice of strings (one per row).
type Sprite struct {
	Frames [][]string
	Color  string
}

// Frame returns the frame at index (wraps around).
func (s *Sprite) Frame(idx int) []string {
	if len(s.Frames) == 0 {
		return nil
	}
	return s.Frames[idx%len(s.Frames)]
}

// HPColor returns a color that fades from the base color toward red
// based on the HP ratio (1.0 = full health, 0.0 = dead).
func HPColor(baseColor string, hpRatio float64) string {
	if hpRatio >= 1.0 {
		return baseColor
	}
	if hpRatio <= 0.0 {
		return "#ff0000"
	}
	if len(baseColor) != 7 {
		return baseColor
	}

	// Parse base color
	br := hexVal(baseColor[1])<<4 + hexVal(baseColor[2])
	bg := hexVal(baseColor[3])<<4 + hexVal(baseColor[4])
	bb := hexVal(baseColor[5])<<4 + hexVal(baseColor[6])

	// Target red: #ff0000
	tr, tg, tb := 255, 0, 0

	// Lerp toward red as HP drops
	t := 1.0 - hpRatio
	r := br + int(float64(tr-br)*t)
	g := bg + int(float64(tg-bg)*t)
	b := bb + int(float64(tb-bb)*t)

	return "#" + hexByte(r) + hexByte(g) + hexByte(b)
}

// All sprites are defined here as multi-char ASCII art.
var Sprites = map[string]*Sprite{
	"player": {
		Frames: [][]string{
			{" O ", "/|\\", "/ \\"}, // idle
			{" O ", "/|\\", "| |"},  // walk
		},
		Color: "#00ff00",
	},
	"rat": {
		Frames: [][]string{
			{"   ", "o--", " ''"},
		},
		Color: "#8B4513",
	},
	"skeleton": {
		Frames: [][]string{
			{"._.", "/X\\", "| |"},
			{"._.", "/X\\", "/ \\"},
		},
		Color: "#cccccc",
	},
	"bat": {
		Frames: [][]string{
			{"/V\\", "\\_/", "   "},
			{"/v\\", "/ \\", "   "},
		},
		Color: "#9966cc",
	},
	"slime": {
		Frames: [][]string{
			{"___", "(o_o)", "~~~"},
		},
		Color: "#33cc33",
	},
	"ghost": {
		Frames: [][]string{
			{"___", "(O.O)", "^^^"},
			{"___", "(o.o)", "^^^"},
		},
		Color: "#aaaaff",
	},
	"minotaur": {
		Frames: [][]string{
			{"(\\=/)", " |X| ", " / \\ "},
		},
		Color: "#cc4400",
	},
	"lich": {
		Frames: [][]string{
			{"{@_@}", " |#| ", " /|\\ "},
		},
		Color: "#8800cc",
	},
	"dragon": {
		Frames: [][]string{
			{"<{O=O}>", " /|=|\\ ", "_/||||\\_"},
		},
		Color: "#ff2200",
	},
}
