package item

import "github.com/df-mc/dragonfly/server/world"

// TurtleShell are items that are used for brewing or as a helmet to give the player the Water Breathing
// status effect.
type TurtleShell struct{}

// Use handles the using of a turtle shell to auto-equip it in an armour slot.
func (t TurtleShell) Use(_ *world.World, user User, _ *UseContext) bool {
	if armoured, ok := user.(Armoured); ok {
		currentEquipped := armoured.Armour().Helmet()

		right, left := user.HeldItems()
		armoured.Armour().SetHelmet(right)
		user.SetHeldItems(currentEquipped, left)
	}
	return false
}

// DurabilityInfo ...
func (t TurtleShell) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 276,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// MaxCount always returns 1.
func (t TurtleShell) MaxCount() int {
	return 1
}

// DefencePoints ...
func (t TurtleShell) DefencePoints() float64 {
	return 2.0
}

// KnockBackResistance ...
func (t TurtleShell) KnockBackResistance() float64 {
	return 0.0
}

// Helmet ...
func (t TurtleShell) Helmet() bool {
	return true
}

// EncodeItem ...
func (t TurtleShell) EncodeItem() (name string, meta int16) {
	return "minecraft:turtle_helmet", 0
}
