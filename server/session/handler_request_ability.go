package session

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// RequestAbilityHandler handles the RequestAbility packet.
type RequestAbilityHandler struct{}

// Handle ...
func (a RequestAbilityHandler) Handle(p packet.Packet, s *Session, tx *world.Tx, c Controllable) error {
	pk := p.(*packet.RequestAbility)
	if pk.Ability == packet.AbilityFlying {
		if !c.GameMode().AllowsFlying() {
			s.log.Debug("process packet: RequestAbility: flying flag enabled while unable to fly")
			s.sendAbilities(c)
			return nil
		}
		c.StartFlying()
	}
	return nil
}
