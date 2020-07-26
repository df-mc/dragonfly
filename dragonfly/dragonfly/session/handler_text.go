package session

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// TextHandler handles the Text packet.
type TextHandler struct{}

// Handle ...
func (TextHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.Text)

	if pk.TextType != packet.TextTypeChat {
		return fmt.Errorf("TextType should always be Chat (%v), but got %v", packet.TextTypeChat, pk.TextType)
	}
	if pk.SourceName != s.conn.IdentityData().DisplayName {
		return fmt.Errorf("SourceName must be equal to DisplayName")
	}
	if pk.XUID != s.conn.IdentityData().XUID {
		return fmt.Errorf("XUID must be equal to player's XUID")
	}
	s.c.Chat(pk.Message)
	return nil
}
