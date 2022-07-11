package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// handleBeaconPayment handles the selection of effects in a beacon and the removal of the item used to pay
// for those effects.
func (h *ItemStackRequestHandler) handleBeaconPayment(a *protocol.BeaconPaymentStackRequestAction, s *Session) error {
	slot := protocol.StackRequestSlotInfo{
		ContainerID: containerBeacon,
		Slot:        0x1b,
	}
	// First check if there actually is a beacon opened.
	if !s.containerOpened.Load() {
		return fmt.Errorf("no beacon container opened")
	}
	pos := s.openedPos.Load()
	beacon, ok := s.c.World().Block(pos).(block.Beacon)
	if !ok {
		return fmt.Errorf("no beacon container opened")
	}

	// Check if the item present in the beacon slot is valid.
	payment, _ := h.itemInSlot(slot, s)
	if payable, ok := payment.Item().(item.BeaconPayment); !ok || !payable.PayableForBeacon() {
		return fmt.Errorf("item %#v in beacon slot cannot be used as payment", payment)
	}

	// Check if the effects are valid and allowed for the beacon's level.
	if !h.validBeaconEffect(a.PrimaryEffect, beacon) {
		return fmt.Errorf("primary effect selected is not allowed: %v for level %v", a.PrimaryEffect, beacon.Level())
	} else if !h.validBeaconEffect(a.SecondaryEffect, beacon) || (beacon.Level() < 4 && a.SecondaryEffect != 0) {
		return fmt.Errorf("secondary effect selected is not allowed: %v for level %v", a.SecondaryEffect, beacon.Level())
	}

	primary, pOk := effect.ByID(int(a.PrimaryEffect))
	secondary, sOk := effect.ByID(int(a.SecondaryEffect))
	if pOk {
		beacon.Primary = primary.(effect.LastingType)
	}
	if sOk {
		beacon.Secondary = secondary.(effect.LastingType)
	}
	s.c.World().SetBlock(pos, beacon, nil)

	// The client will send a Destroy action after this action, but we can't rely on that because the client
	// could just not send it.
	// We just ignore the next Destroy action and set the item to air here.
	h.setItemInSlot(slot, item.Stack{}, s)
	h.ignoreDestroy = true
	return nil
}

// validBeaconEffect checks if the ID passed is a valid beacon effect.
func (h *ItemStackRequestHandler) validBeaconEffect(id int32, beacon block.Beacon) bool {
	switch id {
	case 1, 3:
		return beacon.Level() >= 1
	case 8, 11:
		return beacon.Level() >= 2
	case 5:
		return beacon.Level() >= 3
	case 10:
		return beacon.Level() >= 4
	case 0:
		return true
	}
	return false
}
