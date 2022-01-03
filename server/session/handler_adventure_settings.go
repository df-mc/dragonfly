package session

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// AdventureSettingsHandler handles the AdventureSettings packet.
type AdventureSettingsHandler struct{}

// Handle ...
func (a AdventureSettingsHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.AdventureSettings)

	if pk.PlayerUniqueID != selfEntityRuntimeID {
		return ErrSelfRuntimeID
	}

	mode := s.c.GameMode()
	if pk.Flags&packet.AdventureFlagFlying != 0 {
		if !mode.AllowsFlying() {
			s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: flying flag enabled while not being able to fly\n", s.conn.RemoteAddr(), s.c.Name())
			return nil
		}
		s.c.StartFlying()
	}
	if pk.Flags&packet.AdventureFlagAllowFlight != 0 && !mode.AllowsFlying() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: allow flight flag enabled while not being able to fly\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	if pk.Flags&packet.AdventureFlagNoClip != 0 && !mode.HasCollision() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: no clip flag enabled while not being able to no clip\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	if pk.Flags&packet.AdventureFlagWorldImmutable != 0 && mode.AllowsEditing() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: world immutable flag enabled while being able to edit the world\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	if pk.Flags&packet.AdventureFlagMuted != 0 && mode.Visible() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: muted flag enabled while visible\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	if (pk.ActionPermissions&packet.ActionPermissionMine != 0 || pk.ActionPermissions&packet.ActionPermissionBuild != 0) && !mode.AllowsEditing() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: mine or build permission enabled while not being able to edit the world\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	if (pk.ActionPermissions&packet.ActionPermissionDoorsAndSwitches != 0 || pk.ActionPermissions&packet.ActionPermissionOpenContainers != 0 || pk.ActionPermissions&packet.ActionPermissionAttackPlayers != 0 || pk.ActionPermissions&packet.ActionPermissionAttackMobs != 0) && !mode.AllowsInteraction() {
		s.log.Debugf("failed processing packet from %v (%v): AdventureSettings: doors and switches, open containers, attack players, or attack mobs permissions enabled without being able to interact with the world\n", s.conn.RemoteAddr(), s.c.Name())
		return nil
	}
	return nil
}
