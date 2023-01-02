package inventory

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
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
func NewArmour(f func(slot int, before, after item.Stack)) *Armour {
	inv := New(4, f)
	inv.canAdd = canAddArmour
	return &Armour{inv: inv}
}

// canAddArmour checks if the item passed can be worn as armour in the slot passed.
func canAddArmour(s item.Stack, slot int) bool {
	if s.Empty() {
		return true
	}
	switch slot {
	case 0:
		if h, ok := s.Item().(item.HelmetType); ok {
			return h.Helmet()
		}
	case 1:
		if c, ok := s.Item().(item.ChestplateType); ok {
			return c.Chestplate()
		}
	case 2:
		if l, ok := s.Item().(item.LeggingsType); ok {
			return l.Leggings()
		}
	case 3:
		if b, ok := s.Item().(item.BootsType); ok {
			return b.Boots()
		}
	}
	return false
}

// Set sets all individual pieces of armour in one go. It is equivalent to calling SetHelmet, SetChestplate, SetLeggings
// and SetBoots sequentially.
func (a *Armour) Set(helmet, chestplate, leggings, boots item.Stack) {
	a.SetHelmet(helmet)
	a.SetChestplate(chestplate)
	a.SetLeggings(leggings)
	a.SetBoots(boots)
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

// Slots returns all items (including) air of the armour inventory in the order of helmet, chestplate, leggings,
// boots.
func (a *Armour) Slots() []item.Stack {
	return a.inv.Slots()
}

// Items returns a slice of all non-empty armour items equipped.
func (a *Armour) Items() []item.Stack {
	return a.inv.Items()
}

// Clear clears the armour inventory, removing all items currently present.
func (a *Armour) Clear() []item.Stack {
	return a.inv.Clear()
}

// String converts the armour to a readable string representation.
func (a *Armour) String() string {
	return fmt.Sprintf("(helmet: %v, chestplate: %v, leggings: %v, boots: %v)", a.Helmet(), a.Chestplate(), a.Leggings(), a.Boots())
}

// Inventory returns the underlying Inventory instance.
func (a *Armour) Inventory() *Inventory {
	return a.inv
}

// Handle assigns a Handler to an Armour inventory so that its methods are called for the respective events. Nil may be
// passed to set the default NopHandler.
// Handle is the equivalent of calling (*Armour).Inventory().Handle.
func (a *Armour) Handle(h Handler) {
	a.inv.Handle(h)
}

// Close closes the armour inventory, removing the slot change function.
func (a *Armour) Close() error {
	return a.inv.Close()
}
