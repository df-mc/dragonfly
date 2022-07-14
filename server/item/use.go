package item

// UseContext is passed to every item Use methods. It may be used to subtract items or to deal damage to them
// after the action is complete.
type UseContext struct {
	// Damage is the amount of damage that should be dealt to the item as a result of using it.
	Damage int
	// CountSub is how much of the count should be subtracted after using the item.
	CountSub int
	// IgnoreBBox specifies if placing the item should ignore the BBox of the player placing this. This is the case for
	// items such as cocoa beans.
	IgnoreBBox bool
	// NewItem is the item that is added after the item is used. If the player no longer has an item in the
	// hand, it'll be added there.
	NewItem Stack
	// ConsumedItems contains a list of items that were consumed in the process of using the item.
	ConsumedItems []Stack
	// NewItemSurvivalOnly will add any new items only in survival mode.
	NewItemSurvivalOnly bool

	// FirstFunc returns the first item in the context holder's inventory if found. The second return value describes
	// whether the item was found. The comparable function is used to compare the item to the given item.
	FirstFunc func(comparable func(Stack) bool) (Stack, bool)

	// SwapHeldWithArmour holds a function that swaps the item currently held by a User with armour slot i.
	SwapHeldWithArmour func(i int)
}

// Consume consumes the provided item when the context is handled.
func (ctx *UseContext) Consume(s Stack) {
	ctx.ConsumedItems = append(ctx.ConsumedItems, s)
}

// DamageItem damages the item used by d points.
func (ctx *UseContext) DamageItem(d int) { ctx.Damage += d }

// SubtractFromCount subtracts d from the count of the item stack used.
func (ctx *UseContext) SubtractFromCount(d int) { ctx.CountSub += d }
