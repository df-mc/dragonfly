package session

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"strings"
	"unicode/utf8"
)

type BlockActorDataHandler struct{}

func (b BlockActorDataHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.BlockActorData)
	if id, ok := pk.NBTData["id"]; ok {
		x, y, z := int(pk.Position.X()), int(pk.Position.Y()), int(pk.Position.Z())
		switch id {

		case "Sign":
			blockPosition := cube.Pos{x, y, z}
			if v, ok := s.c.World().Block(blockPosition).(block.WoodSign); ok {

				if pkText, ok := pk.NBTData["Text"]; ok {
					if utf8.ValidString(pkText.(string)) {
						text := strings.TrimRight(pkText.(string), "\n")
						if len(text) >= 100 {
							text = text[:100]
						}
						v.Text = text
					}
				}

				if pkTextOwner, ok := pk.NBTData["TextOwner"]; ok {
					v.TextOwner = pkTextOwner.(string)
				}

				if pkSignTextColor, ok := pk.NBTData["SignTextColor"]; ok {
					if pkSignTextColor.(int32) == 0 {
						pkSignTextColor = int32(-16777216)
					}
					v.SignTextColor = pkSignTextColor.(int32)
				}
				s.c.World().SetBlock(blockPosition, v)
			}
			break
		}
	}
	return nil
}
