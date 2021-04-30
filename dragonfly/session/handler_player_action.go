package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/dragonfly/block"
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PlayerActionHandler handles the PlayerAction packet.
type PlayerActionHandler struct{}

// Handle ...
func (*PlayerActionHandler) Handle(p packet.Packet, s *Session) (err error) {
	pk := p.(*packet.PlayerAction)
	err, _ = handlePlayerAction(pk.ActionType, pk.BlockFace, pk.BlockPosition, pk.EntityRuntimeID, s)
	return
}

// handlePlayerAction handles an action performed by a player, found in packet.PlayerAction and packet.PlayerAuthInput.
func handlePlayerAction(action int32, face int32, pos protocol.BlockPos, entityRuntimeID uint64, s *Session) (error, bool) {
	if entityRuntimeID != selfEntityRuntimeID {
		return ErrSelfRuntimeID, false
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
	case protocol.PlayerActionAbortBreak:
		s.c.AbortBreaking()
	case protocol.PlayerActionStopBreak, protocol.PlayerActionPredictDestroyBlock:
		s.c.FinishBreaking()
	case protocol.PlayerActionCrackBreak, protocol.PlayerActionContinueDestroyBlock:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		newPos := cube.Pos{int(pos[0]), int(pos[1]), int(pos[2])}
		breakingPos, ok := s.c.BreakingPosition()

		if s.c.Breaking() && ok && newPos.Vec3().ApproxEqual(breakingPos.Vec3()) {
			s.c.ContinueBreaking(cube.Face(face))
		} else {
			return handlePlayerAction(protocol.PlayerActionStartBreak, face, pos, entityRuntimeID, s)
		}
	case protocol.PlayerActionStartBreak:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		targetPos := cube.Pos{int(pos[0]), int(pos[1]), int(pos[2])}

		// The client sends a start break action even in cases where it shouldn't. (attempting to break an item with a sword in creative, extinguishing fire, etc)
		// Not sure if there is a better way to handle this.
		if (s.c.GameMode() == world.GameModeCreative{}) {
			held, _ := s.c.HeldItems()
			if _, ok := s.c.World().Block(targetPos.Side(cube.Face(face))).(block.Fire); !ok {
				if _, ok = held.Item().(item.Sword); ok {
					break
				}
			} else {
				s.c.StartBreaking(targetPos, cube.Face(face))
				defer s.c.AbortBreaking()
				return nil, true
			}
		}

		s.c.StartBreaking(targetPos, cube.Face(face))
	case protocol.PlayerActionStartBuildingBlock:
		// Don't do anything for this action.
	case protocol.PlayerActionCreativePlayerDestroyBlock:
		// Don't do anything for this action.
	default:
		return fmt.Errorf("unhandled ActionType %v", action), false
	}
	return nil, false
}
