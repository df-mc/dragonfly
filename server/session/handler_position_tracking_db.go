package session

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
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

// Handle responds with the tracked position, or marks the target as unavailable.
func (*PositionTrackingDBHandler) Handle(p packet.Packet, s *Session, tx *world.Tx, _ Controllable) error {
	pk, ok := p.(*packet.PositionTrackingDBClientRequest)
	if !ok {
		return fmt.Errorf("expected *packet.PositionTrackingDBClientRequest, got %T", p)
	}
	if pk.RequestAction != packet.PositionTrackingDBRequestActionQuery {
		return fmt.Errorf("unknown position tracking request action %d", pk.RequestAction)
	}
	pos, dim, found := tx.World().TrackedPosition(pk.TrackingID)
	if found {
		tx.World().ObservePositionTracking(pk.TrackingID)
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

func positionTrackingPayload(handle int32, pos cube.Pos, dim int, status byte) map[string]any {
	return map[string]any{
		"version": byte(positionTrackingPayloadVersion),
		"dim":     int32(dim),
		"id":      fmt.Sprintf("0x%08x", handle),
		"pos":     []int32{int32(pos[0]), int32(pos[1]), int32(pos[2])},
		"status":  status,
	}
}
