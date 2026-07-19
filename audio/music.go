package audio

import "math/rand"

// MusicState controls the mood of the music.
type MusicState int

const (
	MusicExplore MusicState = iota
	MusicCombat
	MusicBoss
)

// Scale defines a set of note intervals for melody generation.
type Scale struct {
	Notes []string
}

var (
	MinorPentatonic = Scale{Notes: []string{"A", "C", "D", "E", "G"}}
	Dissonant       = Scale{Notes: []string{"A", "A#", "C", "D#", "E", "G#"}}
)

// MusicEngine generates procedural music.
type MusicEngine struct {
	engine       *Engine
	rng          *rand.Rand
	state        MusicState
	targetState  MusicState
	bpm          float64
	nextBeatTime float64
	beatIndex    int
	barCount     int
	scale        Scale
	melodyOctave int
	bassOctave   int
	playing      bool
}

// NewMusicEngine creates a music generator.
func NewMusicEngine(engine *Engine) *MusicEngine {
	return &MusicEngine{
		engine:       engine,
		rng:          rand.New(rand.NewSource(42)),
		state:        MusicExplore,
		targetState:  MusicExplore,
		bpm:          85,
		scale:        MinorPentatonic,
		melodyOctave: 3,
		bassOctave:   2,
	}
}

// SetState requests a music state transition.
func (m *MusicEngine) SetState(s MusicState) {
	m.targetState = s
}

// Update should be called each frame. Schedules notes ahead of time.
func (m *MusicEngine) Update(currentTime float64) {
	if !m.playing {
		m.nextBeatTime = currentTime + 0.5
		m.playing = true
	}

	// Transition state at bar boundaries
	if m.beatIndex%16 == 0 && m.state != m.targetState {
		m.state = m.targetState
		m.applyState()
	}

	// Schedule beats that fall within the next 200ms
	beatDuration := 60.0 / m.bpm

	for m.nextBeatTime < currentTime+0.2 {
		m.scheduleBeat(m.nextBeatTime, beatDuration)
		m.nextBeatTime += beatDuration
		m.beatIndex++
		if m.beatIndex%16 == 0 {
			m.barCount++
		}
	}
}

func (m *MusicEngine) applyState() {
	switch m.state {
	case MusicExplore:
		m.bpm = 85
		m.scale = MinorPentatonic
	case MusicCombat:
		m.bpm = 120
		m.scale = MinorPentatonic
	case MusicBoss:
		m.bpm = 140
		m.scale = Dissonant
	}
}

func (m *MusicEngine) scheduleBeat(t, dur float64) {
	step := m.beatIndex % 16

	// --- Melody (sparse in explore, dense in combat) ---
	playMelody := false
	switch m.state {
	case MusicExplore:
		playMelody = step%4 == 0 || (step%4 == 2 && m.rng.Intn(3) == 0)
	case MusicCombat:
		playMelody = step%2 == 0 || m.rng.Intn(3) == 0
	case MusicBoss:
		playMelody = true
	}

	if playMelody {
		note := m.pickNote()
		freq := NoteFreq(note, m.melodyOctave)
		vol := 0.08
		if m.state == MusicBoss {
			vol = 0.12
		}
		m.engine.PlayTone(freq, t, dur*0.8, "square", vol)
	}

	// --- Bass: root note on beats 0, 4, 8, 12 ---
	if step%4 == 0 && len(m.scale.Notes) > 0 {
		bassNote := m.scale.Notes[0]
		freq := NoteFreq(bassNote, m.bassOctave)
		vol := 0.10
		m.engine.PlayTone(freq, t, dur*2, "sawtooth", vol)
	}

	// --- Percussion ---
	switch m.state {
	case MusicCombat, MusicBoss:
		// Kick on 0, 4, 8, 12
		if step%4 == 0 {
			m.engine.PlayTone(55, t, 0.1, "sine", 0.15)
		}
		// Hihat on every other beat
		if step%2 == 0 {
			m.engine.PlayNoise(t, 0.05, 0.04)
		}
		// Snare-ish on 4, 12
		if step == 4 || step == 12 {
			m.engine.PlayNoise(t, 0.1, 0.08)
		}
	case MusicExplore:
		// Very sparse percussion
		if step == 0 {
			m.engine.PlayTone(55, t, 0.08, "sine", 0.06)
		}
		if step == 8 {
			m.engine.PlayNoise(t, 0.04, 0.02)
		}
	}

	// Vary pattern every 4 bars
	if step == 15 && m.barCount%4 == 3 {
		m.melodyOctave = 3 + m.rng.Intn(2)
	}
}

func (m *MusicEngine) pickNote() string {
	notes := m.scale.Notes
	if len(notes) == 0 {
		return "A"
	}

	// Weight toward root and fifth
	weights := make([]int, len(notes))
	for i := range weights {
		weights[i] = 1
	}
	weights[0] = 3 // root
	if len(weights) > 3 {
		weights[3] = 2 // ~fifth
	}

	total := 0
	for _, w := range weights {
		total += w
	}
	if total == 0 {
		return notes[0]
	}
	roll := m.rng.Intn(total)
	for i, w := range weights {
		roll -= w
		if roll < 0 {
			return notes[i]
		}
	}
	return notes[0]
}
