package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type BlockActorDataHandler struct {}

func (b BlockActorDataHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.BlockActorData)
	x, y, z := int(pk.Position.X()), int(pk.Position.Y()), int(pk.Position.Z())
	switch pk.NBTData["id"]{

	case "Sign":
		blockPosition := cube.Pos{x, y, z}
		if v, ok := s.c.World().Block(blockPosition).(block.WoodSign); ok {
			v.SignTextColor = pk.NBTData["SignTextColor"].(int32)
			v.Text = pk.NBTData["Text"].(string)
			v.TextOwner = pk.NBTData["TextOwner"].(string)
			fmt.Printf("pk x y z%v\n", blockPosition)
			s.c.World().SetBlock(blockPosition, v)
		}
		break
	}
	return nil
}
