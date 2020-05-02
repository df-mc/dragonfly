package inventory

import (
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
)

// Armour represents an inventory for armour. It has 4 slots, one for a helmet, chestplate, leggings and
// boots respectively. NewArmour() must be used to create a valid armour inventory.
// Armour inventories, like normal Inventories, are safe for concurrent usage.
type Armour struct {
	inv *Inventory
}

// NewArmour returns an armour inventory that is ready to be used. The zero value of an inventory.Armour is
// not valid for usage.
// The function passed is called when a slot is changed. It may be nil to not call anything.
func NewArmour(f func(slot int, item item.Stack)) *Armour {
	return &Armour{
		inv: New(4, f),
	}
}

// SetHelmet sets the item stack passed as the helmet in the inventory.
func (a *Armour) SetHelmet(helmet item.Stack) {
	_ = a.inv.SetItem(0, helmet)
}

// Helmet returns the item stack set as helmet in the inventory.
func (a *Armour) Helmet() item.Stack {
	i, _ := a.inv.Item(0)
	return i
}

// SetChestplate sets the item stack passed as the chestplate in the inventory.
func (a *Armour) SetChestplate(chestplate item.Stack) {
	_ = a.inv.SetItem(1, chestplate)
}

// Chestplate returns the item stack set as chestplate in the inventory.
func (a *Armour) Chestplate() item.Stack {
	i, _ := a.inv.Item(1)
	return i
}

// SetLeggings sets the item stack passed as the leggings in the inventory.
func (a *Armour) SetLeggings(leggings item.Stack) {
	_ = a.inv.SetItem(2, leggings)
}

// Leggings returns the item stack set as leggings in the inventory.
func (a *Armour) Leggings() item.Stack {
	i, _ := a.inv.Item(2)
	return i
}

// SetBoots sets the item stack passed as the boots in the inventory.
func (a *Armour) SetBoots(boots item.Stack) {
	_ = a.inv.SetItem(3, boots)
}

// Boots returns the item stack set as boots in the inventory.
func (a *Armour) Boots() item.Stack {
	i, _ := a.inv.Item(3)
	return i
}

// All returns all items (including) air of the armour inventory in the order of helmet, chestplate, leggings,
// boots.
func (a *Armour) All() []item.Stack {
	return a.inv.All()
}

// Clear clears the armour inventory, removing all items currently present.
func (a *Armour) Clear() {
	a.inv.Clear()
}

// String converts the armour to a readable string representation.
func (a *Armour) String() string {
	return fmt.Sprintf("{helmet: %v, chestplate: %v, leggings: %v, boots: %v}", a.Helmet(), a.Chestplate(), a.Leggings(), a.Boots())
}

// Inv returns the underlying Inventory instance.
func (a *Armour) Inv() *Inventory {
	return a.inv
}

// Close closes the armour inventory, removing the slot change function.
func (a *Armour) Close() error {
	return a.inv.Close()
}
