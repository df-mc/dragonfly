package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player/dialogue"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// NPCRequestHandler handles the NPCRequest packet.
type NPCRequestHandler struct {
	dialogue        dialogue.Dialogue
	entityRuntimeID uint64
}

// Handle ...
func (h *NPCRequestHandler) Handle(p packet.Packet, s *Session, tx *world.Tx, c Controllable) error {
	pk := p.(*packet.NPCRequest)
	if pk.RequestType == packet.NPCRequestActionExecuteAction {
		if err := h.dialogue.Submit(uint(pk.ActionType), c, tx); err != nil {
			return fmt.Errorf("error submitting dialogue: %w", err)
		}
	} else if pk.RequestType == packet.NPCRequestActionExecuteClosingCommands {
		h.dialogue.Close(c, tx)
	}
	return nil
}
