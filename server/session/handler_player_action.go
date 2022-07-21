package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
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
		return errSelfRuntimeID
	}
	switch action {
	case protocol.PlayerActionRespawn:
		// Don't do anything for these actions.
	case protocol.PlayerActionDimensionChangeDone:
		if s.switchingWorld.CAS(true, false) {
			s.chunkLoader.Reset()
			s.changeDimension(int32(s.c.World().Dimension().EncodeDimension()), true)
		}
	case protocol.PlayerActionStartBreak, protocol.PlayerActionContinueDestroyBlock:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		s.breakingPos = cube.Pos{int(pos[0]), int(pos[1]), int(pos[2])}
		s.c.StartBreaking(s.breakingPos, cube.Face(face))
	case protocol.PlayerActionAbortBreak:
		s.c.AbortBreaking()
	case protocol.PlayerActionPredictDestroyBlock, protocol.PlayerActionStopBreak:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)
		s.c.FinishBreaking()
	case protocol.PlayerActionCrackBreak:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		newPos := cube.Pos{int(pos[0]), int(pos[1]), int(pos[2])}

		// Sometimes no new position will be sent using a StartBreak action, so we need to detect a change in the
		// block to be broken by comparing positions.
		if newPos != s.breakingPos {
			s.breakingPos = newPos
			s.c.StartBreaking(newPos, cube.Face(face))
			return nil
		}
		s.c.ContinueBreaking(cube.Face(face))
	case protocol.PlayerActionStartItemUseOn, protocol.PlayerActionStopItemUseOn:
		// TODO: Properly utilize these actions.
	case protocol.PlayerActionStartBuildingBlock:
		// Don't do anything for this action.
	case protocol.PlayerActionCreativePlayerDestroyBlock:
		// Don't do anything for this action.
	default:
		return fmt.Errorf("unhandled ActionType %v", action)
	}
	return nil
}
