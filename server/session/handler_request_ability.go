package session

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// RequestAbilityHandler handles the RequestAbility packet.
type RequestAbilityHandler struct{}

// Handle ...
func (a RequestAbilityHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.RequestAbility)
	if pk.Ability == packet.AbilityFlying {
		if !s.c.GameMode().AllowsFlying() {
			s.log.Debugf("failed processing packet from %v (%v): RequestAbility: flying flag enabled while not being able to fly\n", s.conn.RemoteAddr(), s.c.Name())
			s.sendAbilities()
			return nil
		}
		s.c.StartFlying()
	}
	return nil
}
