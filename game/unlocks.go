package game

import "syscall/js"

// LoadoutID identifies a starting loadout.
type LoadoutID int

const (
	LoadoutDefault LoadoutID = iota
	LoadoutWarrior           // Unlocked by beating floor 5
	LoadoutScout             // Unlocked by beating floor 10
	LoadoutGambler           // Unlocked by beating floor 15
)

// Loadout defines a starting configuration.
type Loadout struct {
	ID     LoadoutID
	Name   string
	Desc   string
	Stats  Stats
	Weapon *Item // starting weapon, nil for none
}

var AllLoadouts = []Loadout{
	{
		ID:    LoadoutDefault,
		Name:  "Adventurer",
		Desc:  "Balanced stats, no gear",
		Stats: Stats{STR: 3, DEX: 3, VIT: 3, LCK: 2, Level: 1},
	},
	{
		ID:    LoadoutWarrior,
		Name:  "Warrior",
		Desc:  "High STR/VIT, starts with a sword",
		Stats: Stats{STR: 5, DEX: 2, VIT: 4, LCK: 1, Level: 1},
		Weapon: &Item{
			Name: "Rusty Longsword", Type: ItemWeapon,
			Rarity: RarityCommon, BonusSTR: 2,
		},
	},
	{
		ID:    LoadoutScout,
		Name:  "Scout",
		Desc:  "High DEX/LCK, fast and lucky",
		Stats: Stats{STR: 2, DEX: 6, VIT: 2, LCK: 3, Level: 1},
	},
	{
		ID:    LoadoutGambler,
		Name:  "Gambler",
		Desc:  "Max LCK, random starting stats",
		Stats: Stats{STR: 2, DEX: 2, VIT: 2, LCK: 6, Level: 1},
	},
}

// UnlockFloor returns the floor that must be beaten to unlock a loadout.
func UnlockFloor(id LoadoutID) int {
	switch id {
	case LoadoutWarrior:
		return 5
	case LoadoutScout:
		return 10
	case LoadoutGambler:
		return 15
	default:
		return 0
	}
}

const storageKey = "gorl_unlocks"

// UnlockManager handles persistent unlocks via localStorage.
type UnlockManager struct {
	Unlocked map[LoadoutID]bool
}

// NewUnlockManager loads unlock state from localStorage.
func NewUnlockManager() *UnlockManager {
	um := &UnlockManager{
		Unlocked: map[LoadoutID]bool{
			LoadoutDefault: true, // always available
		},
	}
	um.load()
	return um
}

// IsUnlocked returns true if the loadout is available.
func (um *UnlockManager) IsUnlocked(id LoadoutID) bool {
	return um.Unlocked[id]
}

// CheckUnlocks grants any loadouts earned by reaching the given floor.
func (um *UnlockManager) CheckUnlocks(floorReached int) []string {
	var newUnlocks []string
	for _, lo := range AllLoadouts {
		if um.Unlocked[lo.ID] {
			continue
		}
		req := UnlockFloor(lo.ID)
		if req > 0 && floorReached >= req {
			um.Unlocked[lo.ID] = true
			newUnlocks = append(newUnlocks, lo.Name)
		}
	}
	if len(newUnlocks) > 0 {
		um.save()
	}
	return newUnlocks
}

// UnlockedLoadouts returns all available loadouts.
func (um *UnlockManager) UnlockedLoadouts() []Loadout {
	var result []Loadout
	for _, lo := range AllLoadouts {
		if um.Unlocked[lo.ID] {
			result = append(result, lo)
		}
	}
	return result
}

func (um *UnlockManager) save() {
	storage := js.Global().Get("localStorage")
	if storage.IsUndefined() || storage.IsNull() {
		return
	}

	// Simple format: comma-separated IDs
	val := ""
	for id := range um.Unlocked {
		if val != "" {
			val += ","
		}
		val += itoa(int(id))
	}
	storage.Call("setItem", storageKey, val)
}

func (um *UnlockManager) load() {
	storage := js.Global().Get("localStorage")
	if storage.IsUndefined() || storage.IsNull() {
		return
	}

	raw := storage.Call("getItem", storageKey)
	if raw.IsNull() || raw.IsUndefined() {
		return
	}

	str := raw.String()
	if str == "" {
		return
	}

	// Parse comma-separated IDs
	num := 0
	for i := 0; i <= len(str); i++ {
		if i == len(str) || str[i] == ',' {
			um.Unlocked[LoadoutID(num)] = true
			num = 0
		} else if str[i] >= '0' && str[i] <= '9' {
			num = num*10 + int(str[i]-'0')
		}
	}
}
