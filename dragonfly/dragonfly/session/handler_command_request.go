package session

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// CommandRequestHandler handles the CommandRequest packet.
type CommandRequestHandler struct {
	origin protocol.CommandOrigin
}

// Handle ...
func (h *CommandRequestHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.CommandRequest)
	if pk.Internal {
		return fmt.Errorf("command request packet must never have the internal field set to true")
	}

	h.origin = pk.CommandOrigin
	s.c.ExecuteCommand(pk.CommandLine)
	return nil
}
