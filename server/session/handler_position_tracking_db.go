package session

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	positionTrackingPayloadVersion  = 1
	positionTrackingStatusTracked   = 0
	positionTrackingStatusDestroyed = 2
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
	resp := &packet.PositionTrackingDBServerBroadcast{TrackingID: pk.TrackingID, Payload: map[string]any{}}
	switch pos, dim, status := tx.World().TrackedPosition(pk.TrackingID); status {
	case world.PositionTracked:
		resp.BroadcastAction = packet.PositionTrackingDBBroadcastActionUpdate
		resp.Payload = positionTrackingPayload(pk.TrackingID, pos, dim, positionTrackingStatusTracked)
	case world.PositionDestroyed:
		resp.BroadcastAction = packet.PositionTrackingDBBroadcastActionDestroy
		resp.Payload = positionTrackingPayload(pk.TrackingID, cube.Pos{}, -1, positionTrackingStatusDestroyed)
	default:
		resp.BroadcastAction = packet.PositionTrackingDBBroadcastActionNotFound
	}
	s.writePacket(resp)
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
