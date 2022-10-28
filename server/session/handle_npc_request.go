package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player/dialogue"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync"
)

// NpcRequestHandler handles the NpcRequest packet.
type NpcRequestHandler struct {
	mu        sync.Mutex
	dialogues map[string]dialogue.Dialogue
}

// Handle ...
func (h *NpcRequestHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.NPCRequest)
	h.mu.Lock()
	d, ok := h.dialogues[s.c.XUID()]
	h.mu.Unlock()
	if !ok {
		return fmt.Errorf("no dialogue menu for player with xuid %v", s.c.XUID())
	}
	m := d.Menu()
	switch pk.RequestType {
	case packet.NPCRequestActionExecuteAction:
		buttons := m.Buttons()
		index := int(pk.ActionType)
		if index >= len(buttons) {
			return fmt.Errorf("error submitting dialogue, button index points to inexistent button: %v (only %v buttons present)", index, len(buttons))
		}
		d.Submit(s.Controllable(), buttons[index])
		// We close the dialogue because if we don't close it here and the api implementor forgets to close it the
		// client permanently stuck in the dialogue UI being unable to close it or submit a button.
		s.Controllable().CloseDialogue(d)
	case packet.NPCRequestActionExecuteClosingCommands:
		if c, ok := d.(dialogue.Closer); ok {
			c.Close(s.Controllable())
		}
		h.mu.Lock()
		delete(h.dialogues, s.c.XUID())
		h.mu.Unlock()
	}
	return nil
}
