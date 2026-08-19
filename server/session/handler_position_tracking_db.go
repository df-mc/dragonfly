package session

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	positionTrackingPayloadVersion = 1
	positionTrackingStatusTracked  = 0
	positionTrackingStatusMissing  = 2
)

// PositionTrackingDBHandler handles client queries for lodestone compass targets.
type PositionTrackingDBHandler struct{}

// Handle ...
func (*PositionTrackingDBHandler) Handle(p packet.Packet, s *Session, tx *world.Tx, _ Controllable) error {
	pk, ok := p.(*packet.PositionTrackingDBClientRequest)
	if !ok {
		return fmt.Errorf("expected *packet.PositionTrackingDBClientRequest, got %T", p)
	}
	if pk.RequestAction != packet.PositionTrackingDBRequestActionQuery {
		return fmt.Errorf("unknown position tracking request action %d", pk.RequestAction)
	}
	pos, dim, found := cube.Pos{}, 0, false
	if s.holdsTrackingCompass(pk.TrackingID) {
		pos, dim, found = tx.World().TrackedPosition(pk.TrackingID)
	}
	action, status := byte(packet.PositionTrackingDBBroadcastActionUpdate), byte(positionTrackingStatusTracked)
	if !found {
		action, status = packet.PositionTrackingDBBroadcastActionNotFound, positionTrackingStatusMissing
	}
	s.writePacket(&packet.PositionTrackingDBServerBroadcast{
		BroadcastAction: action,
		TrackingID:      pk.TrackingID,
		Payload:         positionTrackingPayload(pk.TrackingID, pos, dim, status),
	})
	return nil
}

// holdsTrackingCompass reports whether a compass linked to handle sits anywhere the client is currently
// drawing for this player. Queries for handles it holds no compass for are refused, so lodestone positions
// cannot be found by walking the handle space.
func (s *Session) holdsTrackingCompass(handle int32) bool {
	hasHandle := func(stack item.Stack) bool {
		compass, ok := stack.Item().(item.Compass)
		return ok && compass.TrackingHandle == handle
	}
	// ui covers the cursor and crafting grid, and openedWindow the container the player is looking into,
	// including an ender chest, which is stored there while open.
	inventories := []*inventory.Inventory{s.inv, s.offHand, s.ui, s.openedWindow.Load()}
	for _, inv := range inventories {
		if inv == nil {
			continue
		}
		if _, ok := inv.FirstFunc(hasHandle); ok {
			return true
		}
	}
	return false
}

func positionTrackingPayload(handle int32, pos cube.Pos, dim int, status byte) map[string]any {
	return map[string]any{
		"version": byte(positionTrackingPayloadVersion),
		"dim":     int32(dim),
		"id":      fmt.Sprintf("0x%08x", handle),
		"pos":     []int32{int32(pos[0]), int32(pos[1]), int32(pos[2])},
		"status":  status,
	}
}
