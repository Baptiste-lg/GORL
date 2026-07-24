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
			{" /\\", "{||}", " /\\"}, // idle - hooded cloak
			{" /\\", "{||}", " ||"},  // walk
		},
		Color: "#44ff88",
	},
	"rat": {
		Frames: [][]string{
			{";\\ ", "(\")", " ''"},
		},
		Color: "#cc8844",
	},
	"skeleton": {
		Frames: [][]string{
			{".^.", "/+\\", "| |"},
			{".^.", "/+\\", "/ \\"},
		},
		Color: "#eeeedd",
	},
	"bat": {
		Frames: [][]string{
			{"^.^", "/W\\", "   "},
			{"v.v", "\\W/", "   "},
		},
		Color: "#cc99ff",
	},
	"slime": {
		Frames: [][]string{
			{" _ ", "(~)", "~~~"},
			{"._.", "(o~)", " ~~"},
		},
		Color: "#55ee55",
	},
	"ghost": {
		Frames: [][]string{
			{".-.", "(O)", "^^^"},
			{".-.", "(o)", "~~~"},
		},
		Color: "#ccccff",
	},
	"minotaur": {
		Frames: [][]string{
			{"(\\=/)", " |X| ", "_/ \\_"},
		},
		Color: "#ff7733",
	},
	"lich": {
		Frames: [][]string{
			{"{@_@}", " ]#[ ", " /|\\ "},
			{"{o_o}", " ]#[ ", " /|\\ "},
		},
		Color: "#cc55ff",
	},
	"dragon": {
		Frames: [][]string{
			{"<{O=O}>", " /|=|\\ ", "_/|||\\_"},
			{"<{o=o}>", " /|=|\\ ", "_/ | \\_"},
		},
		Color: "#ff5533",
	},
}
