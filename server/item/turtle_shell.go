package item

import "github.com/df-mc/dragonfly/server/world"

// TurtleShell are items that are used for brewing or as a helmet to give the player the Water Breathing
// status effect.
type TurtleShell struct{}

// Use handles the using of a turtle shell to auto-equip it in an armour slot.
func (TurtleShell) Use(_ *world.Tx, _ User, ctx *UseContext) bool {
	ctx.SwapHeldWithArmour(0)
	return false
}

func (TurtleShell) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 276,
		BrokenItem:    simpleItem(Stack{}),
	}
}

func (TurtleShell) RepairableBy(i Stack) bool {
	_, ok := i.Item().(Scute)
	return ok
}

// MaxCount always returns 1.
func (TurtleShell) MaxCount() int {
	return 1
}

func (TurtleShell) DefencePoints() float64 {
	return 2
}

func (TurtleShell) Toughness() float64 {
	return 0
}

func (TurtleShell) KnockBackResistance() float64 {
	return 0
}

func (TurtleShell) EnchantmentValue() int {
	return 9
}

func (TurtleShell) Helmet() bool {
	return true
}

func (TurtleShell) EncodeItem() (name string, meta int16) {
	return "minecraft:turtle_helmet", 0
}
