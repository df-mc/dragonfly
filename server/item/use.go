package item

// UseContext is passed to every item Use methods. It may be used to subtract items or to deal damage to them
// after the action is complete.
type UseContext struct {
	// IgnoreAABB specifies if placing the item should ignore the AABB of the player placing this. This is the case for
	// items such as cocoa beans.
	IgnoreAABB bool
	// Damage is the amount of damage that should be dealt to the item as a result of using it.
	Damage int
	// CountSub is how much of the count should be subtracted after using the item.
	CountSub int
	// NewItem is the item that is added after the item is used. If the player no longer has an item in the
	// hand, it'll be added there.
	NewItem Stack
	// NewItemSurvivalOnly will add any new items only in survival mode.
	NewItemSurvivalOnly bool

	// SwapHeldWithArmour holds a function that swaps the item currently held by a User with armour slot i.
	SwapHeldWithArmour func(i int)
}

// DamageItem damages the item used by d points.
func (ctx *UseContext) DamageItem(d int) {
	ctx.Damage += d
}

// SubtractFromCount subtracts d from the count of the item stack used.
func (ctx *UseContext) SubtractFromCount(d int) {
	ctx.CountSub += d
}
