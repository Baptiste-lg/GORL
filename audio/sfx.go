package audio

// SFX provides game sound effects.
type SFX struct {
	engine *Engine
}

// NewSFX creates a sound effects player.
func NewSFX(engine *Engine) *SFX {
	return &SFX{engine: engine}
}

// Hit plays a short noise burst for melee hits.
func (s *SFX) Hit() {
	t := s.engine.CurrentTime()
	s.engine.PlayNoise(t, 0.08, 0.12)
	s.engine.PlayTone(200, t, 0.06, "square", 0.08)
}

// CritHit plays a louder, lower hit sound.
func (s *SFX) CritHit() {
	t := s.engine.CurrentTime()
	s.engine.PlayNoise(t, 0.12, 0.18)
	s.engine.PlayTone(150, t, 0.1, "square", 0.12)
	s.engine.PlayTone(100, t+0.05, 0.1, "sawtooth", 0.08)
}

// Pickup plays an ascending arpeggio for item collection.
func (s *SFX) Pickup() {
	t := s.engine.CurrentTime()
	s.engine.PlayTone(523, t, 0.08, "square", 0.06)
	s.engine.PlayTone(659, t+0.08, 0.08, "square", 0.06)
	s.engine.PlayTone(784, t+0.16, 0.12, "square", 0.06)
}

// LevelUp plays a fanfare (major chord arpeggio).
func (s *SFX) LevelUp() {
	t := s.engine.CurrentTime()
	s.engine.PlayTone(523, t, 0.15, "square", 0.08)
	s.engine.PlayTone(659, t+0.15, 0.15, "square", 0.08)
	s.engine.PlayTone(784, t+0.30, 0.15, "square", 0.08)
	s.engine.PlayTone(1047, t+0.45, 0.30, "square", 0.10)
}

// DoorOpen plays a low descending creak.
func (s *SFX) DoorOpen() {
	t := s.engine.CurrentTime()
	s.engine.PlayTone(180, t, 0.15, "sawtooth", 0.05)
	s.engine.PlayTone(120, t+0.1, 0.15, "sawtooth", 0.04)
}

// Footstep plays a very quiet click.
func (s *SFX) Footstep() {
	t := s.engine.CurrentTime()
	s.engine.PlayNoise(t, 0.02, 0.015)
}

// Miss plays a quiet whoosh.
func (s *SFX) Miss() {
	t := s.engine.CurrentTime()
	s.engine.PlayNoise(t, 0.1, 0.04)
}

// Stairs plays an ascending tone for floor transition.
func (s *SFX) Stairs() {
	t := s.engine.CurrentTime()
	s.engine.PlayTone(220, t, 0.1, "sine", 0.06)
	s.engine.PlayTone(330, t+0.1, 0.1, "sine", 0.06)
	s.engine.PlayTone(440, t+0.2, 0.2, "sine", 0.08)
}

// Death plays a descending tone.
func (s *SFX) Death() {
	t := s.engine.CurrentTime()
	s.engine.PlayTone(440, t, 0.2, "sawtooth", 0.10)
	s.engine.PlayTone(330, t+0.2, 0.2, "sawtooth", 0.08)
	s.engine.PlayTone(220, t+0.4, 0.3, "sawtooth", 0.06)
	s.engine.PlayTone(110, t+0.7, 0.5, "sawtooth", 0.04)
}
