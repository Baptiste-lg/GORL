package game

import "math/rand"

// ShrineType identifies the kind of shrine.
type ShrineType int

const (
	ShrineBlood   ShrineType = iota // Sacrifice HP for a stat boost
	ShrineFortune                   // Gamble: blessing or curse
	ShrineHealing                   // Full heal (one use)
)

// Shrine is an interactable object placed in special rooms.
type Shrine struct {
	X, Y int
	Type ShrineType
	Used bool
}

// Glyph returns the display character.
func (s *Shrine) Glyph() string {
	if s.Used {
		return "_"
	}
	switch s.Type {
	case ShrineBlood:
		return "&"
	case ShrineFortune:
		return "$"
	case ShrineHealing:
		return "+"
	}
	return "&"
}

// Color returns the display color.
func (s *Shrine) Color() string {
	if s.Used {
		return "#444444"
	}
	switch s.Type {
	case ShrineBlood:
		return "#cc2222"
	case ShrineFortune:
		return "#ccaa00"
	case ShrineHealing:
		return "#22cc22"
	}
	return "#cc2222"
}

// Name returns the shrine's display name.
func (s *Shrine) Name() string {
	switch s.Type {
	case ShrineBlood:
		return "Blood Altar"
	case ShrineFortune:
		return "Fortune Shrine"
	case ShrineHealing:
		return "Healing Fountain"
	}
	return "Shrine"
}

// ShrineResult holds the outcome of using a shrine.
type ShrineResult struct {
	Message string
	Color   string
}

// UseShrine activates a shrine's effect on the player.
func UseShrine(s *Shrine, p *Player, rng *rand.Rand) ShrineResult {
	if s.Used {
		return ShrineResult{Message: "Already used", Color: "#444444"}
	}
	s.Used = true

	switch s.Type {
	case ShrineBlood:
		return useShrineBlood(p, rng)
	case ShrineFortune:
		return useShrineFortune(p, rng)
	case ShrineHealing:
		return useShrineHealing(p)
	}
	return ShrineResult{}
}

func useShrineBlood(p *Player, rng *rand.Rand) ShrineResult {
	// Sacrifice 30% of current HP for +2 to a random stat
	cost := p.Stats.HP * 30 / 100
	if cost < 1 {
		cost = 1
	}
	p.Entity.TakeDamage(cost)

	stats := []string{"STR", "DEX", "VIT", "LCK"}
	pick := rng.Intn(len(stats))
	switch pick {
	case 0:
		p.Stats.STR += 2
	case 1:
		p.Stats.DEX += 2
	case 2:
		p.Stats.VIT += 2
	case 3:
		p.Stats.LCK += 2
	}

	return ShrineResult{
		Message: "-" + intStr(cost) + " HP, +2 " + stats[pick],
		Color:   "#cc2222",
	}
}

func useShrineFortune(p *Player, rng *rand.Rand) ShrineResult {
	// 50% blessing, 50% curse
	if rng.Intn(2) == 0 {
		// Blessing: +3 to a random stat
		stats := []string{"STR", "DEX", "VIT", "LCK"}
		pick := rng.Intn(len(stats))
		switch pick {
		case 0:
			p.Stats.STR += 3
		case 1:
			p.Stats.DEX += 3
		case 2:
			p.Stats.VIT += 3
		case 3:
			p.Stats.LCK += 3
		}
		return ShrineResult{
			Message: "Blessed! +3 " + stats[pick],
			Color:   "#FFD700",
		}
	}

	// Curse: -2 to a random stat (min 1)
	stats := []string{"STR", "DEX", "VIT", "LCK"}
	pick := rng.Intn(len(stats))
	switch pick {
	case 0:
		p.Stats.STR -= 2
		if p.Stats.STR < 1 {
			p.Stats.STR = 1
		}
	case 1:
		p.Stats.DEX -= 2
		if p.Stats.DEX < 1 {
			p.Stats.DEX = 1
		}
	case 2:
		p.Stats.VIT -= 2
		if p.Stats.VIT < 1 {
			p.Stats.VIT = 1
		}
	case 3:
		p.Stats.LCK -= 2
		if p.Stats.LCK < 1 {
			p.Stats.LCK = 1
		}
	}
	return ShrineResult{
		Message: "Cursed! -2 " + stats[pick],
		Color:   "#aa22aa",
	}
}

func useShrineHealing(p *Player) ShrineResult {
	maxHP := p.EffectiveStats().MaxHP()
	healed := maxHP - p.Stats.HP
	p.Stats.HP = maxHP
	return ShrineResult{
		Message: "Healed +" + intStr(healed) + " HP",
		Color:   "#22cc22",
	}
}

// SpawnShrines places shrines in some rooms.
func SpawnShrines(rooms []*Room, floor int, rng *rand.Rand) []*Shrine {
	var shrines []*Shrine

	for i, room := range rooms {
		// Skip spawn room and stairs room (last)
		if i == 0 || i == len(rooms)-1 {
			continue
		}

		// ~15% chance per room to contain a shrine
		if rng.Intn(100) >= 15 {
			continue
		}

		shrineType := ShrineType(rng.Intn(3))
		shrines = append(shrines, &Shrine{
			X:    room.X + room.W/2,
			Y:    room.Y + room.H/2,
			Type: shrineType,
		})
	}

	return shrines
}

func intStr(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 5)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
