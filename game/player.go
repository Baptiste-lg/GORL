package game

// StatusEffect is a temporary buff on the player.
type StatusEffect struct {
	Name      string
	Remaining float64 // seconds left
	Kind      ScrollKind
	Magnitude int // for reversible stat buffs (STR/VIT bonus to remove on expiry)
}

// Player wraps an Entity with player-specific state.
type Player struct {
	*Entity
	Inventory    *Inventory
	Active       *ActiveItem
	Synergies    []SynergyID
	Gold         int
	MoveCooldown float64 // seconds remaining until next move
	LastCombat   float64 // seconds since last combat (for HP regen)
	RegenTimer   float64 // accumulator for passive regen ticks
	Effects      []StatusEffect
}

// RecalcSynergies updates the cached synergy list based on current equipment and active item.
func (p *Player) RecalcSynergies() {
	p.Synergies = DetectSynergies(p.Inventory.Weapon, p.Inventory.Armor, p.Active)
}

// NewPlayer creates a player entity at the given position.
func NewPlayer(x, y int) *Player {
	stats := DefaultPlayerStats()
	return &Player{
		Entity:     NewEntity(x, y, "player", stats),
		Inventory:  NewInventory(),
		LastCombat: 999,
	}
}

// CanMove returns true if the movement cooldown has elapsed.
func (p *Player) CanMove() bool {
	return p.MoveCooldown <= 0
}

// ResetMoveCooldown sets the cooldown based on effective DEX.
func (p *Player) ResetMoveCooldown() {
	cd := p.EffectiveStats().MoveCooldownMS() / 1000.0
	// Speed scroll halves cooldown
	if p.HasEffect(ScrollSpeed) {
		cd /= 2
	}
	// Weapon affix: Swift (-20% move cooldown)
	if p.WeaponHasAffix(AffixSwift) {
		cd *= 0.80
	}
	p.MoveCooldown = cd
}

// EffectiveStats returns stats with equipment bonuses and armor affixes applied.
func (p *Player) EffectiveStats() Stats {
	s := p.Stats
	str, dex, vit, lck := p.Inventory.EquipBonuses()
	s.STR += str
	s.DEX += dex
	s.VIT += vit
	s.LCK += lck

	// Armor affixes
	if a := p.Inventory.Armor; a != nil {
		if HasAffix(a.Affixes, AffixFortified) {
			s.VIT += s.VIT / 4 // +25% VIT
		}
	}
	return s
}

// WeaponHasAffix returns true if the equipped weapon has the given affix.
func (p *Player) WeaponHasAffix(id AffixID) bool {
	if w := p.Inventory.Weapon; w != nil {
		return HasAffix(w.Affixes, id)
	}
	return false
}

// ArmorHasAffix returns true if the equipped armor has the given affix.
func (p *Player) ArmorHasAffix(id AffixID) bool {
	if a := p.Inventory.Armor; a != nil {
		return HasAffix(a.Affixes, id)
	}
	return false
}

// UseItem uses a consumable item at the given inventory index.
func (p *Player) UseItem(idx int) bool {
	if idx < 0 || idx >= len(p.Inventory.Items) {
		return false
	}
	item := p.Inventory.Items[idx]

	switch item.Type {
	case ItemPotion:
		healAmt := int(PotionHealPercent(item.Rarity) * float64(p.EffectiveStats().MaxHP()))
		p.Heal(healAmt)
		p.Inventory.Remove(idx)
		return true
	case ItemScroll:
		duration := 10.0
		if item.ScrollKind == ScrollShield {
			duration = 15.0
		}
		p.AddEffect(StatusEffect{
			Name:      item.Name,
			Remaining: duration,
			Kind:      item.ScrollKind,
		})
		p.Inventory.Remove(idx)
		return true
	}
	return false
}

// AddEffect adds a status effect (refreshes if same kind exists).
func (p *Player) AddEffect(eff StatusEffect) {
	for i := range p.Effects {
		if p.Effects[i].Kind == eff.Kind {
			p.Effects[i].Remaining = eff.Remaining
			return
		}
	}
	p.Effects = append(p.Effects, eff)
}

// HasEffect returns true if the player has an active effect of the given kind.
func (p *Player) HasEffect(kind ScrollKind) bool {
	for _, e := range p.Effects {
		if e.Kind == kind && e.Remaining > 0 {
			return true
		}
	}
	return false
}

// Update ticks player timers and status effects.
func (p *Player) Update(dt float64) {
	if p.MoveCooldown > 0 {
		p.MoveCooldown -= dt
	}
	p.LastCombat += dt

	// Tick status effects
	alive := p.Effects[:0]
	for i := range p.Effects {
		p.Effects[i].Remaining -= dt
		if p.Effects[i].Remaining > 0 {
			alive = append(alive, p.Effects[i])
		} else {
			// Reverse stat buffs on expiry
			p.onEffectExpired(p.Effects[i])
		}
	}
	p.Effects = alive
}

func (p *Player) onEffectExpired(eff StatusEffect) {
	if eff.Magnitude == 0 {
		return
	}
	switch eff.Kind {
	case ScrollKind(101): // War Cry: reverse STR
		p.Stats.STR -= eff.Magnitude
		if p.Stats.STR < 1 {
			p.Stats.STR = 1
		}
	case ScrollKind(104): // Shield Wall: reverse VIT
		p.Stats.VIT -= eff.Magnitude
		if p.Stats.VIT < 1 {
			p.Stats.VIT = 1
		}
	}
}
