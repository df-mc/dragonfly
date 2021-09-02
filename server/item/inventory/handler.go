package inventory

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
)

type Handler interface {
	HandleTake(ctx *event.Context, slot int, it item.Stack)

	HandlePlace(ctx *event.Context, slot int, it item.Stack)

	HandleDrop(ctx *event.Context, slot int, it item.Stack)
}

// NopHandler is an implementation of Handler that does not run any code in any of its methods. It is the default
// Handler of an Inventory.
type NopHandler struct{}

func (NopHandler) HandleTake(*event.Context, int, item.Stack)  {}
func (NopHandler) HandlePlace(*event.Context, int, item.Stack) {}
func (NopHandler) HandleDrop(*event.Context, int, item.Stack)  {}
