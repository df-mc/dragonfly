package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PlayerSkinHandler handles the PlayerSkin packet.
type PlayerSkinHandler struct{}

// Handle ...
func (PlayerSkinHandler) Handle(p packet.Packet, _ *Session, _ *world.Tx, c Controllable) error {
	pk := p.(*packet.PlayerSkin)

	playerSkin, err := protocolToSkin(pk.Skin)
	if err != nil {
		return fmt.Errorf("error decoding skin: %w", err)
	}

	c.SetSkin(playerSkin)

	return nil
}
