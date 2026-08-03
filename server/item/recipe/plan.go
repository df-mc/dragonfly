package recipe

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
)

// CraftingPlan represents the result of a crafting calculation performed before session-side slot application.
type CraftingPlan struct {
	// Inputs contains the concrete input stacks consumed by the craft.
	Inputs []item.Stack
	// Results contains the output stacks produced by the craft.
	Results []item.Stack
	// Changes contains the inventory slot updates needed to apply the craft.
	Changes []CraftingSlotChange
}

// CraftingSlotChange represents a deferred inventory slot mutation for a crafting operation.
type CraftingSlotChange struct {
	// Inventory is the inventory containing the slot to update.
	Inventory *inventory.Inventory
	// Slot is the slot index within Inventory that should be updated.
	Slot int
	// Stack is the resulting stack that should be written into Slot.
	Stack item.Stack
}
