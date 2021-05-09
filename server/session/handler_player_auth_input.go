package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PlayerAuthInputHandler handles the PlayerAuthInput packet.
type PlayerAuthInputHandler struct{}

// Handle ...
func (h PlayerAuthInputHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.PlayerAuthInput)
	if err := h.handleMovement(pk, s); err != nil {
		return err
	}
	return h.handleActions(pk, s)
}

// handleMovement handles the movement part of the packet.PlayerAuthInput.
func (h PlayerAuthInputHandler) handleMovement(pk *packet.PlayerAuthInput, s *Session) error {
	pk.Position = pk.Position.Sub(mgl32.Vec3{0, 1.62}) // Subtract the base offset of players from the pos.

	newPos := vec32To64(pk.Position)
	deltaPos, deltaYaw, deltaPitch := newPos.Sub(s.c.Position()), float64(pk.Yaw)-s.c.Yaw(), float64(pk.Pitch)-s.c.Pitch()
	if mgl64.FloatEqual(deltaPos.Len(), 0) && mgl64.FloatEqual(deltaYaw, 0) && mgl64.FloatEqual(deltaPitch, 0) {
		// The PlayerAuthInput packet is sent every tick, so don't do anything if the position and rotation
		// were unchanged.
		return nil
	}

	s.teleportMu.Lock()
	teleportPos := s.teleportPos
	s.teleportMu.Unlock()
	if teleportPos != nil {
		if newPos.Sub(*teleportPos).Len() > 0.5 {
			// The player has moved before it received the teleport packet. Ignore this movement entirely and
			// wait for the client to sync itself back to the server. Once we get a movement that is close
			// enough to the teleport position, we'll allow the player to move around again.
			return nil
		}
		s.teleportMu.Lock()
		s.teleportPos = nil
		s.teleportMu.Unlock()
	}

	_, submergedBefore := s.c.World().Liquid(cube.PosFromVec3(s.c.Position().Add(mgl64.Vec3{0, s.c.EyeHeight()})))

	s.c.Move(deltaPos)
	s.c.Rotate(deltaYaw, deltaPitch)

	_, submergedAfter := s.c.World().Liquid(cube.PosFromVec3(s.c.Position().Add(mgl64.Vec3{0, s.c.EyeHeight()})))

	if submergedBefore != submergedAfter {
		// Player wasn't either breathing before and no longer isn't, or wasn't breathing before and now is,
		// so send the updated metadata.
		s.ViewEntityState(s.c, s.c.State())
	}

	s.chunkLoader.Move(s.c.Position())
	s.writePacket(&packet.NetworkChunkPublisherUpdate{
		Position: protocol.BlockPos{int32(pk.Position[0]), int32(pk.Position[1]), int32(pk.Position[2])},
		Radius:   uint32(s.chunkRadius) << 4,
	})
	return nil
}

// handleInteractions handles the actions with the world that are present in the PlayerAuthInput packet.
func (h PlayerAuthInputHandler) handleActions(pk *packet.PlayerAuthInput, s *Session) error {
	if pk.InputData&packet.InputFlagPerformItemInteraction != 0 {
		if err := h.handleUseItemData(pk.ItemInteractionData, s); err != nil {
			return err
		}
	}
	if pk.InputData&packet.InputFlagPerformItemStackRequest != 0 {
		// God knows what this is for.
		s.log.Debugf("PlayerAuthInput: unexpected item stack request: %#v\n", pk.ItemStackRequest)
	}
	if pk.InputData&packet.InputFlagPerformBlockActions != 0 {
		if err := h.handleBlockActions(pk.BlockActions, s); err != nil {
			return err
		}
	}
	return nil
}

// handleUseItemData handles the protocol.UseItemTransactionData found in a packet.PlayerAuthInput.
func (h PlayerAuthInputHandler) handleUseItemData(data protocol.UseItemTransactionData, s *Session) error {
	held, _ := s.c.HeldItems()
	if !held.Equal(stackToItem(data.HeldItem.Stack)) {
		s.log.Debugf("failed processing item interaction from %v (%v): PlayerAuthInput: actual held and client held item mismatch", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	pos := cube.Pos{int(data.BlockPosition[0]), int(data.BlockPosition[1]), int(data.BlockPosition[2])}
	s.swingingArm.Store(true)
	defer s.swingingArm.Store(false)

	// Seems like this is only used for breaking blocks at the moment.
	switch data.ActionType {
	case protocol.UseItemActionBreakBlock:
		s.c.BreakBlock(pos)
	default:
		return fmt.Errorf("unhandled UseItem ActionType for PlayerAuthInput packet %v", data.ActionType)
	}
	return nil
}

// handleBlockActions handles a slice of protocol.PlayerBlockAction present in a PlayerAuthInput packet.
func (h PlayerAuthInputHandler) handleBlockActions(a []protocol.PlayerBlockAction, s *Session) error {
	for _, action := range a {
		if err := handlePlayerAction(action.Action, action.Face, action.BlockPos, selfEntityRuntimeID, s); err != nil {
			return err
		}
	}
	return nil
}
