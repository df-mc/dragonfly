package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PlayerActionHandler handles the PlayerAction packet.
type PlayerActionHandler struct{}

// Handle ...
func (*PlayerActionHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.PlayerAction)

	return handlePlayerAction(pk.ActionType, pk.BlockFace, pk.BlockPosition, pk.EntityRuntimeID, s)
}

// handlePlayerAction handles an action performed by a player, found in packet.PlayerAction and packet.PlayerAuthInput.
func handlePlayerAction(action int32, face int32, pos protocol.BlockPos, entityRuntimeID uint64, s *Session) error {
	if entityRuntimeID != selfEntityRuntimeID {
		return ErrSelfRuntimeID
	}
	switch action {
	case protocol.PlayerActionRespawn:
		// Don't do anything for this action.
	case protocol.PlayerActionJump:
		if s.c.Sprinting() {
			s.c.Exhaust(0.2)
		} else {
			s.c.Exhaust(0.05)
		}
	case protocol.PlayerActionStartSprint:
		s.c.StartSprinting()
	case protocol.PlayerActionStopSprint:
		s.c.StopSprinting()
	case protocol.PlayerActionStartSneak:
		s.c.StartSneaking()
	case protocol.PlayerActionStopSneak:
		s.c.StopSneaking()
	case protocol.PlayerActionStartSwimming:
		if _, ok := s.c.World().Liquid(cube.BlockPosFromVec3(s.c.Position().Add(mgl64.Vec3{0, s.c.EyeHeight()}))); ok {
			s.c.StartSwimming()
		}
	case protocol.PlayerActionStopSwimming:
		s.c.StopSwimming()
	case protocol.PlayerActionContinueDestroyBlock:
		fallthrough
	case protocol.PlayerActionStartBreak:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		s.c.StartBreaking(cube.Pos{int(pos[0]), int(pos[1]), int(pos[2])}, cube.Face(face))
	case protocol.PlayerActionAbortBreak:
		s.c.AbortBreaking()
	case protocol.PlayerActionPredictDestroyBlock:
		fallthrough
	case protocol.PlayerActionStopBreak:
		s.c.FinishBreaking()
	case protocol.PlayerActionCrackBreak:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		s.c.ContinueBreaking(cube.Face(face))
	case protocol.PlayerActionStartBuildingBlock:
		// Don't do anything for this action.
	case protocol.PlayerActionCreativePlayerDestroyBlock:
		// Don't do anything for this action.
	default:
		return fmt.Errorf("unhandled ActionType %v", action)
	}
	return nil
}
