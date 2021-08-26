package session

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type LevelSoundEventHandler struct{}

func (l LevelSoundEventHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.LevelSoundEvent)
	if pk.SoundType == packet.SoundEventAttackNoDamage && (s.c.GameMode() != world.GameModeSpectator{}) {
		ctx := event.C()
		s.c.PunchAir(ctx)
		ctx.Continue(func() {
			s.swingingArm.Store(true)
			defer s.swingingArm.Store(false)
			s.c.SwingArm()
			s.c.World().PlaySound(s.c.Position(), sound.Attack{Damage: false})
		})
	}
	return nil
}
