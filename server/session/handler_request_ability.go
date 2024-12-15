package session

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// RequestAbilityHandler handles the RequestAbility packet.
type RequestAbilityHandler struct{}

// Handle ...
func (a RequestAbilityHandler) Handle(p packet.Packet, s *Session, _ *world.Tx, c Controllable) error {
	pk := p.(*packet.RequestAbility)
	if pk.Ability == packet.AbilityFlying {
		if !c.GameMode().AllowsFlying() {
			s.conf.Log.Debug("process packet: RequestAbility: flying flag enabled while unable to fly")
			s.SendAbilities(c)
			return nil
		}
		c.StartFlying()
	}
	return nil
}
