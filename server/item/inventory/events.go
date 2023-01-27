package inventory

import (
    "github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
)

type EventDrop struct {
    Inventory *Inventory
    Slot int
    ItemStack item.Stack
    *event.Context
}

type EventPlace struct {
    Inventory *Inventory
    Slot int
    ItemStack item.Stack
    *event.Context
}

type EventTake struct {
    Inventory *Inventory
    Slot int
    ItemStack item.Stack
    *event.Context
}
