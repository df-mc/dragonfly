package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PlayerActionHandler handles the PlayerAction packet.
type PlayerActionHandler struct{}

// Handle ...
func (*PlayerActionHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.PlayerAction)

	if pk.EntityRuntimeID != selfEntityRuntimeID {
		return ErrSelfRuntimeID
	}
	switch pk.ActionType {
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
		if _, ok := s.c.World().Liquid(world.BlockPosFromVec3(s.c.Position().Add(mgl64.Vec3{0, s.c.EyeHeight()}))); ok {
			s.c.StartSwimming()
		}
	case protocol.PlayerActionStopSwimming:
		s.c.StopSwimming()
	case protocol.PlayerActionStartBreak:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		s.c.StartBreaking(world.BlockPos{int(pk.BlockPosition[0]), int(pk.BlockPosition[1]), int(pk.BlockPosition[2])}, world.Face(pk.BlockFace))
	case protocol.PlayerActionAbortBreak:
		s.c.AbortBreaking()
	case protocol.PlayerActionStopBreak:
		s.c.FinishBreaking()
	case protocol.PlayerActionCrackBreak:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		s.c.ContinueBreaking(world.Face(pk.BlockFace))
	case protocol.PlayerActionStartBuildingBlock:
		// Don't do anything for this action.
	case protocol.PlayerActionCreativePlayerDestroyBlock:
		// Don't do anything for this action.
	default:
		return fmt.Errorf("unhandled ActionType %v", pk.ActionType)
	}
	return nil
}
