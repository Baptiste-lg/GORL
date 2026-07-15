package game

const MaxInventorySlots = 10

// Inventory holds the player's items and equipment.
type Inventory struct {
	Items    []*Item
	Weapon   *Item
	Armor    *Item
	Selected int // cursor position in inventory UI
}

// NewInventory creates an empty inventory.
func NewInventory() *Inventory {
	return &Inventory{}
}

// IsFull returns true if there's no room for more items.
func (inv *Inventory) IsFull() bool {
	return len(inv.Items) >= MaxInventorySlots
}

// Add puts an item in the inventory. Returns false if full.
func (inv *Inventory) Add(item *Item) bool {
	if inv.IsFull() {
		return false
	}
	item.X = 0
	item.Y = 0
	inv.Items = append(inv.Items, item)
	return true
}

// Remove removes an item at the given index and returns it.
func (inv *Inventory) Remove(idx int) *Item {
	if idx < 0 || idx >= len(inv.Items) {
		return nil
	}
	item := inv.Items[idx]
	inv.Items = append(inv.Items[:idx], inv.Items[idx+1:]...)
	if inv.Selected >= len(inv.Items) && inv.Selected > 0 {
		inv.Selected--
	}
	return item
}

// Equip puts an item in its equipment slot. Returns the previously equipped item (or nil).
func (inv *Inventory) Equip(idx int) *Item {
	if idx < 0 || idx >= len(inv.Items) {
		return nil
	}
	item := inv.Items[idx]
	var old *Item

	switch item.Type {
	case ItemWeapon:
		old = inv.Weapon
		inv.Weapon = item
	case ItemArmor:
		old = inv.Armor
		inv.Armor = item
	default:
		return nil // can't equip potions/scrolls
	}

	// Remove from inventory
	inv.Items = append(inv.Items[:idx], inv.Items[idx+1:]...)

	// Put old item back in inventory
	if old != nil {
		inv.Items = append(inv.Items, old)
	}

	if inv.Selected >= len(inv.Items) && inv.Selected > 0 {
		inv.Selected--
	}
	return old
}

// SelectedItem returns the currently highlighted item, or nil.
func (inv *Inventory) SelectedItem() *Item {
	if inv.Selected < 0 || inv.Selected >= len(inv.Items) {
		return nil
	}
	return inv.Items[inv.Selected]
}

// EquipBonuses returns total stat bonuses from equipped items.
func (inv *Inventory) EquipBonuses() (str, dex, vit, lck int) {
	if inv.Weapon != nil {
		str += inv.Weapon.BonusSTR
		dex += inv.Weapon.BonusDEX
		vit += inv.Weapon.BonusVIT
		lck += inv.Weapon.BonusLCK
	}
	if inv.Armor != nil {
		str += inv.Armor.BonusSTR
		dex += inv.Armor.BonusDEX
		vit += inv.Armor.BonusVIT
		lck += inv.Armor.BonusLCK
	}
	return
}
