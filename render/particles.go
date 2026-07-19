package render

// Particle is a floating text effect (damage numbers, MISS, etc).
type Particle struct {
	X, Y    float64 // world tile position
	Text    string
	Color   string
	Life    float64 // remaining seconds
	MaxLife float64
	VY      float64 // vertical velocity (tiles/sec, negative = up)
}

// ParticleSystem manages active particles.
type ParticleSystem struct {
	Particles []*Particle
}

// NewParticleSystem creates an empty particle system.
func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{}
}

// SpawnDamage creates a floating damage number.
func (ps *ParticleSystem) SpawnDamage(x, y, amount int, color string) {
	ps.Particles = append(ps.Particles, &Particle{
		X:       float64(x),
		Y:       float64(y),
		Text:    "-" + intToStr(amount),
		Color:   color,
		Life:    0.8,
		MaxLife: 0.8,
		VY:      -2.0,
	})
}

// SpawnCrit creates a CRIT damage number.
func (ps *ParticleSystem) SpawnCrit(x, y, amount int) {
	ps.Particles = append(ps.Particles, &Particle{
		X:       float64(x),
		Y:       float64(y),
		Text:    "CRIT -" + intToStr(amount),
		Color:   "#FFD700",
		Life:    1.0,
		MaxLife: 1.0,
		VY:      -2.5,
	})
}

// SpawnMiss creates a MISS text.
func (ps *ParticleSystem) SpawnMiss(x, y int) {
	ps.Particles = append(ps.Particles, &Particle{
		X:       float64(x),
		Y:       float64(y),
		Text:    "MISS",
		Color:   "#888888",
		Life:    0.6,
		MaxLife: 0.6,
		VY:      -1.5,
	})
}

// SpawnText creates a generic floating text.
func (ps *ParticleSystem) SpawnText(x, y int, text, color string) {
	ps.Particles = append(ps.Particles, &Particle{
		X:       float64(x),
		Y:       float64(y),
		Text:    text,
		Color:   color,
		Life:    1.0,
		MaxLife: 1.0,
		VY:      -2.0,
	})
}

// Update ticks all particles and removes dead ones.
func (ps *ParticleSystem) Update(dt float64) {
	alive := ps.Particles[:0]
	for _, p := range ps.Particles {
		p.Life -= dt
		p.Y += p.VY * dt
		if p.Life > 0 {
			alive = append(alive, p)
		}
	}
	ps.Particles = alive
}

// Draw renders all particles using the renderer.
func (ps *ParticleSystem) Draw(r *Renderer) {
	for _, p := range ps.Particles {
		// Convert world position to viewport cell
		vx := int(p.X) - r.CamX
		vy := p.Y - float64(r.CamY)

		if vx < 0 || vx >= ViewTilesX || vy < 0 || vy >= float64(ViewTilesY) {
			continue
		}

		col := vx*TileCells + 1
		row := int(vy*float64(TileCells)) + 1

		// Fade alpha based on remaining life
		alpha := p.Life / p.MaxLife
		color := p.Color
		if alpha < 0.5 {
			color = dimParticleColor(p.Color)
		}

		r.DrawText(col, row, p.Text, color)
	}
}

func dimParticleColor(hex string) string {
	if len(hex) != 7 {
		return "#555555"
	}
	r := hexVal(hex[1])<<4 + hexVal(hex[2])
	g := hexVal(hex[3])<<4 + hexVal(hex[4])
	b := hexVal(hex[5])<<4 + hexVal(hex[6])
	r /= 2
	g /= 2
	b /= 2
	return "#" + hexByte(r) + hexByte(g) + hexByte(b)
}
