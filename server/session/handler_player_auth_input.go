package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"math"
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
	yaw, pitch := s.c.Rotation()
	pos := s.c.Position()

	reference := []float64{pitch, yaw, yaw, pos[0], pos[1], pos[2]}
	for i, v := range [...]*float32{&pk.Pitch, &pk.Yaw, &pk.HeadYaw, &pk.Position[0], &pk.Position[1], &pk.Position[2]} {
		f := float64(*v)
		if math.IsNaN(f) || math.IsInf(f, 1) || math.IsInf(f, 0) {
			// Sometimes, the PlayerAuthInput packet is in fact sent with NaN/INF after being teleported (to another
			// world), see #425. For this reason, we don't actually return an error if this happens, because this will
			// result in the player being kicked. Just log it and replace the NaN value with the one we have tracked
			// server-side.
			s.log.Debugf("failed processing packet from %v (%v): %T: must not have nan/inf values, but got %v (%v, %v, %v). assuming server-side values\n", s.conn.RemoteAddr(), s.c.Name(), pk, pk.Position, pk.Pitch, pk.Yaw, pk.HeadYaw)
			*v = float32(reference[i])
		}
	}

	pk.Position = pk.Position.Sub(mgl32.Vec3{0, 1.62}) // Sub the base offset of players from the pos.

	newPos := vec32To64(pk.Position)
	deltaPos, deltaYaw, deltaPitch := newPos.Sub(pos), float64(pk.Yaw)-yaw, float64(pk.Pitch)-pitch
	if mgl64.FloatEqual(deltaPos.Len(), 0) && mgl64.FloatEqual(deltaYaw, 0) && mgl64.FloatEqual(deltaPitch, 0) {
		// The PlayerAuthInput packet is sent every tick, so don't do anything if the position and rotation
		// were unchanged.
		return nil
	}

	if expected := s.teleportPos.Load(); expected != nil {
		if newPos.Sub(*expected).Len() > 1 {
			// The player has moved before it received the teleport packet. Ignore this movement entirely and
			// wait for the client to sync itself back to the server. Once we get a movement that is close
			// enough to the teleport position, we'll allow the player to move around again.
			return nil
		}
		s.teleportPos.Store(nil)
	}

	s.c.Move(deltaPos, deltaYaw, deltaPitch)

	if !mgl64.FloatEqual(deltaPos.Len(), 0) {
		s.chunkLoader.Move(newPos)
		s.writePacket(&packet.NetworkChunkPublisherUpdate{
			Position: protocol.BlockPos{int32(pk.Position[0]), int32(pk.Position[1]), int32(pk.Position[2])},
			Radius:   uint32(s.chunkRadius) << 4,
		})
	}
	return nil
}

// handleActions handles the actions with the world that are present in the PlayerAuthInput packet.
func (h PlayerAuthInputHandler) handleActions(pk *packet.PlayerAuthInput, s *Session) error {
	if pk.InputData&packet.InputFlagPerformItemInteraction != 0 {
		if err := h.handleUseItemData(pk.ItemInteractionData, s); err != nil {
			return err
		}
	}
	if pk.InputData&packet.InputFlagPerformBlockActions != 0 {
		if err := h.handleBlockActions(pk.BlockActions, s); err != nil {
			return err
		}
	}
	h.handleInputFlags(pk.InputData, s)

	if pk.InputData&packet.InputFlagPerformItemStackRequest != 0 {
		s.inTransaction.Store(true)
		defer s.inTransaction.Store(false)

		// As of 1.18 this is now used for sending item stack requests such as when mining a block.
		sh := s.handlers[packet.IDItemStackRequest].(*ItemStackRequestHandler)
		if err := sh.handleRequest(pk.ItemStackRequest, s); err != nil {
			// Item stacks being out of sync isn't uncommon, so don't error. Just debug the error and let the
			// revert do its work.
			s.log.Debugf("failed processing packet from %v (%v): PlayerAuthInput: error resolving item stack request: %v", s.conn.RemoteAddr(), s.c.Name(), err)
		}
	}
	return nil
}

// handleInputFlags handles the toggleable input flags set in a PlayerAuthInput packet.
func (h PlayerAuthInputHandler) handleInputFlags(flags uint64, s *Session) {
	if flags&packet.InputFlagStartSprinting != 0 {
		s.c.StartSprinting()
	}
	if flags&packet.InputFlagStopSprinting != 0 {
		s.c.StopSprinting()
	}
	if flags&packet.InputFlagStartSneaking != 0 {
		s.c.StartSneaking()
	}
	if flags&packet.InputFlagStopSneaking != 0 {
		s.c.StopSneaking()
	}
	if flags&packet.InputFlagStartSwimming != 0 {
		s.c.StartSwimming()
	}
	if flags&packet.InputFlagStopSwimming != 0 {
		s.c.StopSwimming()
	}
	if flags&packet.InputFlagStartJumping != 0 {
		s.c.Jump()
	}
}

// handleUseItemData handles the protocol.UseItemTransactionData found in a packet.PlayerAuthInput.
func (h PlayerAuthInputHandler) handleUseItemData(data protocol.UseItemTransactionData, s *Session) error {
	s.swingingArm.Store(true)
	defer s.swingingArm.Store(false)

	held, _ := s.c.HeldItems()
	if !held.Equal(stackToItem(data.HeldItem.Stack)) {
		s.log.Debugf("failed processing item interaction from %v (%v): PlayerAuthInput: actual held and client held item mismatch", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	pos := cube.Pos{int(data.BlockPosition[0]), int(data.BlockPosition[1]), int(data.BlockPosition[2])}

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
