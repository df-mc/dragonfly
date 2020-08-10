package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
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
	case packet.PlayerActionRespawn:
		// Don't do anything for this action.
	case packet.PlayerActionJump:
		if s.c.Sprinting() {
			s.c.Exhaust(0.2)
		} else {
			s.c.Exhaust(0.05)
		}
	case packet.PlayerActionStartSprint:
		s.c.StartSprinting()
	case packet.PlayerActionStopSprint:
		s.c.StopSprinting()
	case packet.PlayerActionStartSneak:
		s.c.StartSneaking()
	case packet.PlayerActionStopSneak:
		s.c.StopSneaking()
	case packet.PlayerActionStartSwimming:
		if _, ok := s.c.World().Liquid(world.BlockPosFromVec3(s.c.Position().Add(mgl64.Vec3{0, s.c.EyeHeight()}))); ok {
			s.c.StartSwimming()
		}
	case packet.PlayerActionStopSwimming:
		s.c.StopSwimming()
	case packet.PlayerActionStartBreak:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		s.c.StartBreaking(world.BlockPos{int(pk.BlockPosition[0]), int(pk.BlockPosition[1]), int(pk.BlockPosition[2])}, world.Face(pk.BlockFace))
	case packet.PlayerActionAbortBreak:
		s.c.AbortBreaking()
	case packet.PlayerActionStopBreak:
		s.c.FinishBreaking()
	case packet.PlayerActionContinueBreak:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		s.c.ContinueBreaking(world.Face(pk.BlockFace))
	case packet.PlayerActionStartBuildingBlock:
		// Don't do anything for this action.
	default:
		return fmt.Errorf("unhandled ActionType %v", pk.ActionType)
	}
	return nil
}
