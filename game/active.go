package game

// ActiveID identifies an active item type.
type ActiveID int

const (
	ActiveNone        ActiveID = -1
	ActiveDash        ActiveID = iota // Teleport 3 tiles in move direction
	ActiveFireball                    // Deal 15 damage in a line (3 tiles)
	ActiveFreeze                      // Freeze all visible enemies for 2s
	ActiveHealBurst                   // Heal 40% max HP
	ActiveShieldWall                  // +10 VIT for 8 seconds
	ActivePoisonCloud                 // Poison all adjacent enemies
	ActiveBlink                       // Teleport to random explored tile
	ActiveWarCry                      // +5 STR for 6 seconds

	activeCount
)

// ActiveItem is a usable ability in the player's active slot.
type ActiveItem struct {
	ID         ActiveID
	Name       string
	Desc       string
	MaxCharges int // kills needed to charge
	Charges    int // current kill count toward next use
}

// Ready returns true if the active item is fully charged.
func (a *ActiveItem) Ready() bool {
	return a.Charges >= a.MaxCharges
}

// Use consumes charges. Returns false if not ready.
func (a *ActiveItem) Use() bool {
	if !a.Ready() {
		return false
	}
	a.Charges = 0
	return true
}

// AddCharge increments charge count (capped at MaxCharges).
func (a *ActiveItem) AddCharge() {
	if a.Charges < a.MaxCharges {
		a.Charges++
	}
}

// ActiveDef holds definition data for an active item type.
type ActiveDef struct {
	ID         ActiveID
	Name       string
	Desc       string
	MaxCharges int
}

var ActiveDefs = map[ActiveID]ActiveDef{
	ActiveDash:        {ActiveDash, "Dash", "Teleport 3 tiles forward", 3},
	ActiveFireball:    {ActiveFireball, "Fireball", "15 fire damage in a line", 5},
	ActiveFreeze:      {ActiveFreeze, "Freeze", "Freeze visible enemies 2s", 6},
	ActiveHealBurst:   {ActiveHealBurst, "Heal Burst", "Heal 40% max HP", 5},
	ActiveShieldWall:  {ActiveShieldWall, "Shield Wall", "+10 VIT for 8s", 4},
	ActivePoisonCloud: {ActivePoisonCloud, "Poison Cloud", "Poison adjacent foes", 4},
	ActiveBlink:       {ActiveBlink, "Blink", "Teleport to random tile", 3},
	ActiveWarCry:      {ActiveWarCry, "War Cry", "+5 STR for 6s", 5},
}

// NewActiveItem creates an active item from a definition.
func NewActiveItem(id ActiveID) *ActiveItem {
	def, ok := ActiveDefs[id]
	if !ok {
		return nil
	}
	return &ActiveItem{
		ID:         def.ID,
		Name:       def.Name,
		Desc:       def.Desc,
		MaxCharges: def.MaxCharges,
	}
}

var allActiveIDs = []ActiveID{
	ActiveDash, ActiveFireball, ActiveFreeze, ActiveHealBurst,
	ActiveShieldWall, ActivePoisonCloud, ActiveBlink, ActiveWarCry,
}

// RandomActiveID picks a random active item type.
func RandomActiveID(rng interface{ Intn(int) int }) ActiveID {
	return allActiveIDs[rng.Intn(len(allActiveIDs))]
}
