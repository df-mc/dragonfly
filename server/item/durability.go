package item

// Durable represents an item that has durability, and may therefore be broken. Some durable items, when
// broken, create a new item, such as an elytra.
type Durable interface {
	// DurabilityInfo returns info related to the durability of an item.
	DurabilityInfo() DurabilityInfo
}

// DurabilityInfo is the info of a durable item. It includes fields that must be set in order to define
// durability related behaviour.
type DurabilityInfo struct {
	// MaxDurability is the maximum durability that this item may have. This field must be positive for the
	// durability to function properly.
	MaxDurability int
	// BrokenItem is the item created when the item is broken. For most durable items, this is simply an
	// air item.
	BrokenItem func() Stack
	// AttackDurability and BreakDurability are the losses in durability that the item sustains when they are
	// used to do the respective actions.
	AttackDurability, BreakDurability int
}

// Repairable represents a durable item that can be repaired by other items.
type Repairable interface {
	Durable
	RepairableBy(i Stack) bool
}

// simpleItem is a convenience function to return an item stack as BrokenItem.
func simpleItem(i Stack) func() Stack {
	return func() Stack {
		return i
	}
}
