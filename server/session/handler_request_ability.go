package session

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// RequestAbilityHandler handles the RequestAbility packet.
type RequestAbilityHandler struct{}

// Handle ...
func (a RequestAbilityHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.RequestAbility)

	mode := s.c.GameMode()
	if pk.Ability == packet.AbilityFlying {
		if !mode.AllowsFlying() {
			s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: flying flag enabled while not being able to fly\n", s.conn.RemoteAddr(), s.c.Name())
			return nil
		}
		s.c.StartFlying()
	}
	if pk.Ability == packet.AbilityMayFly && !mode.AllowsFlying() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: allow flight flag enabled while not being able to fly\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	if pk.Ability == packet.AbilityNoClip && !mode.HasCollision() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: no clip flag enabled while not being able to no clip\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	if pk.Ability == packet.AbilityWorldBuilder && mode.AllowsEditing() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: world immutable flag enabled while being able to edit the world\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	if pk.Ability == packet.AdventureFlagMuted && mode.Visible() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: muted flag enabled while visible\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	if (pk.Ability == packet.AbilityMine || pk.Ability == packet.AbilityBuild) && !mode.AllowsEditing() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: mine or build permission enabled while not being able to edit the world\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	if (pk.Ability == packet.AbilityDoorsAndSwitches || pk.Ability == packet.AbilityOpenContainers || pk.Ability == packet.AbilityAttackPlayers || pk.Ability == packet.AbilityAttackMobs) && !mode.AllowsInteraction() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: doors and switches, open containers, attack players, or attack mobs permissions enabled without being able to interact with the world\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	return nil
}
