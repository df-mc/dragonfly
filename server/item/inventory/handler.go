package inventory

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
)

// Handler is a type that may be used to handle actions performed on an inventory by a player.
type Handler interface {
	// HandleTake handles an item.Stack being taken from a slot in the inventory. This item might be the whole stack or
	// part of the stack currently present in that slot.
	HandleTake(ctx *event.Context, slot int, it item.Stack)
	// HandlePlace handles an item.Stack being placed in a slot of the inventory. It might either be added to an empty
	// slot or a slot that contains an item of the same type.
	HandlePlace(ctx *event.Context, slot int, it item.Stack)
	// HandleDrop handles the dropping of an item.Stack in a slot out of the inventory.
	HandleDrop(ctx *event.Context, slot int, it item.Stack)
}

// Check to make sure NopHandler implements Handler.
var _ Handler = NopHandler{}

// NopHandler is an implementation of Handler that does not run any code in any of its methods. It is the default
// Handler of an Inventory.
type NopHandler struct{}

func (NopHandler) HandleTake(*event.Context, int, item.Stack)  {}
func (NopHandler) HandlePlace(*event.Context, int, item.Stack) {}
func (NopHandler) HandleDrop(*event.Context, int, item.Stack)  {}
