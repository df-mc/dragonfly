package session

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PlayerActionHandler handles the PlayerAction packet.
type PlayerActionHandler struct{}

// Handle ...
func (*PlayerActionHandler) Handle(p packet.Packet, s *Session, _ *world.Tx, c Controllable) error {
	pk := p.(*packet.PlayerAction)

	return handlePlayerAction(pk.ActionType, pk.BlockFace, pk.BlockPosition, pk.EntityRuntimeID, s, c)
}

// handlePlayerAction handles an action performed by a player, found in packet.PlayerAction and packet.PlayerAuthInput.
func handlePlayerAction(action int32, face int32, pos protocol.BlockPos, entityRuntimeID uint64, s *Session, c Controllable) error {
	if entityRuntimeID != selfEntityRuntimeID {
		return errSelfRuntimeID
	}
	switch action {
	case protocol.PlayerActionStartSleeping, protocol.PlayerActionRespawn, protocol.PlayerActionDimensionChangeDone:
		// Don't do anything for these actions.
	case protocol.PlayerActionStopSleeping:
		c.Wake()
	case protocol.PlayerActionStartBreak:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		s.breakingFace = cube.Face(face)
		c.StartBreaking(blockPos(pos), s.breakingFace)
	case protocol.PlayerActionContinueDestroyBlock:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		s.breakingFace, s.continuedBreak = cube.Face(face), true
		c.ContinueBreakingBlock(blockPos(pos), s.breakingFace)
	case protocol.PlayerActionAbortBreak:
		c.AbortBreaking()
	case protocol.PlayerActionPredictDestroyBlock:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		predicted := blockPos(pos)
		if c.GameMode().CreativeInventory() && s.continuedBreak {
			c.StartBreaking(predicted, cube.Face(face))
		}
		s.predictedBreak = &predicted
	case protocol.PlayerActionStopBreak, protocol.PlayerActionCrackBreak:
		// Only the client's cracking animation stops: the block is broken once the server has made
		// enough progress breaking it.
	case protocol.PlayerActionStartItemUseOn:
		// TODO: Properly utilize these actions.
	case protocol.PlayerActionStopItemUseOn:
		c.ReleaseItem()
	case protocol.PlayerActionInteractWithBlock:
		// Don't do anything for this action.
	case protocol.PlayerActionCreativePlayerDestroyBlock:
		// Don't do anything for this action.
	case protocol.PlayerActionMissedSwing:
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)
		c.PunchAir()
	default:
		return fmt.Errorf("unhandled ActionType %v", action)
	}
	return nil
}

// blockPos converts a protocol.BlockPos to a cube.Pos.
func blockPos(pos protocol.BlockPos) cube.Pos {
	return cube.Pos{int(pos[0]), int(pos[1]), int(pos[2])}
}
