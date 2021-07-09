package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"strings"
	"unicode/utf8"
)

// BlockActorDataHandler handles an incoming BlockActorData packet from the client, sent for some block entities like
// signs when they are edited.
type BlockActorDataHandler struct{}

// Handle ...
func (b BlockActorDataHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.BlockActorData)
	if id, ok := pk.NBTData["id"]; ok {
		pos := cube.Pos{int(pk.Position.X()), int(pk.Position.Y()), int(pk.Position.Z())}
		switch id {
		case "Sign":
			return b.handleSign(pk, pos, s)
		}
		return fmt.Errorf("unhandled block actor data ID %v", id)
	}
	return fmt.Errorf("block actor data without 'id' tag: %v", pk.NBTData)
}

// handleSign handles the BlockActorData packet sent when editing a sign.
func (b BlockActorDataHandler) handleSign(pk *packet.BlockActorData, pos cube.Pos, s *Session) error {
	sign, ok := s.c.World().Block(pos).(block.Sign)
	if !ok {
		return fmt.Errorf("sign block actor data for position without sign %v", pos)
	}
	// TODO: Make sure only the owner of the sign can edit it.

	var text string
	pkText, ok := pk.NBTData["Text"]
	if !ok {
		return fmt.Errorf("sign block actor data had no 'Text' tag")
	}
	if text, ok = pkText.(string); !ok {
		return fmt.Errorf("sign block actor data 'Text' tag was not a string: %#v", pkText)
	}
	text = strings.TrimRight(text, "\n")
	if len(text) > 100 {
		return fmt.Errorf("sign block actor data text was longer than 100 characters")
	}
	if !utf8.ValidString(text) {
		return fmt.Errorf("sign block actor data text was not valid UTF8")
	}
	sign.Text, sign.TextOwner = text, s.conn.IdentityData().XUID
	s.c.World().SetBlock(pos, sign)
	return nil
}
