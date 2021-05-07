package session

import (
	"github.com/df-mc/dragonfly/server/entity/action"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type LevelSoundEventHandler struct{}

func (l LevelSoundEventHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.LevelSoundEvent)

	if pk.SoundType == packet.SoundEventAttackNoDamage {
		s.swingingArm.Store(true)
		defer s.swingingArm.Store(false)

		pos := s.Position()
		for _, v := range s.c.World().Viewers(s.Position()) {
			v.ViewSound(pos, sound.Attack{Damage: false})
			v.ViewEntityAction(s.c, action.SwingArm{})
		}
	}

	return nil
}
