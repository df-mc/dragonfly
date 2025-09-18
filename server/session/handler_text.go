package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"time"
	"unicode/utf8"
)

// TextHandler handles the Text packet.
type TextHandler struct {
	charactersCounter            int
	lastCharacterCounterResetsAt time.Time
}

// Handle ...
func (h *TextHandler) Handle(p packet.Packet, s *Session, _ *world.Tx, c Controllable) error {
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
	c.Chat(pk.Message)

	const (
		// Client's chat input field has a maximum length of 256 characters.
		maxChatCharLength = 256
		// A character can be up to 4 bytes in UTF-8, so we set the maximum byte length to 4 times the maximum
		// character length.
		maxChatByteLength = maxChatCharLength * 4
	)
	// We check the message byte length first as it is O(1), whereas the utf8.RuneCountInString is O(n).
	if len(pk.Message) >= maxChatByteLength {
		return fmt.Errorf("message byte length exceeds maximum of %d bytes", maxChatByteLength)
	}
	characterLen := utf8.RuneCountInString(pk.Message)
	if characterLen > maxChatCharLength {
		return fmt.Errorf("message character length exceeds maximum of %d characters", maxChatCharLength)
	}

	// We limit the number of characters a player can send in chat per tick to prevent spam.
	if time.Since(h.lastCharacterCounterResetsAt) > time.Second/20 {
		h.charactersCounter = 0
		h.lastCharacterCounterResetsAt = time.Now()
	}
	h.charactersCounter += characterLen

	return nil
}
