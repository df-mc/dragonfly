package player

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
)

// armourHandler is an inventory.Handler assigned to a Player's armour inventory. It calls
// Handler.HandleArmourEquip before an item is placed into one of the armour slots, so that equipping may be
// prevented.
type armourHandler struct {
	inventory.NopHandler
	p *Player
}

// HandlePlace calls Handler.HandleArmourEquip for the item being placed. Cancelling the Context passed to it
// cancels the inventory Context too, which stops the item from being placed at all.
func (h armourHandler) HandlePlace(ctx *inventory.Context, slot int, it item.Stack) {
	before, _ := h.p.armour.Inventory().Item(slot)

	c := newContext(h.p)
	if h.p.Handler().HandleArmourEquip(c, slot, before, it); c.Cancelled() {
		ctx.Cancel()
	}
}
