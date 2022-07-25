package item

import "github.com/df-mc/dragonfly/server/world"

// Elytra is a pair of rare wings found in end ships that are the only single-item source of flight in Survival mode.
type Elytra struct{}

// Use handles the using of an elytra to auto-equip it in an armour slot.
func (Elytra) Use(_ *world.World, _ User, ctx *UseContext) bool {
	ctx.SwapHeldWithArmour(1)
	return false
}

// DurabilityInfo ...
func (Elytra) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 433,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// RepairableBy ...
func (Elytra) RepairableBy(i Stack) bool {
	_, ok := i.Item().(PhantomMembrane)
	return ok
}

// MaxCount always returns 1.
func (Elytra) MaxCount() int {
	return 1
}

// Chestplate ...
func (Elytra) Chestplate() bool {
	return true
}

// DefencePoints ...
func (Elytra) DefencePoints() float64 {
	return 0
}

// Toughness ...
func (e Elytra) Toughness() float64 {
	return 0
}

// KnockBackResistance ...
func (e Elytra) KnockBackResistance() float64 {
	return 0
}

// EncodeItem ...
func (Elytra) EncodeItem() (name string, meta int16) {
	return "minecraft:elytra", 0
}
