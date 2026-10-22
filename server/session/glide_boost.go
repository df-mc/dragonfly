package session

import (
	"time"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// viewGlideBoost shows a glide boost of the duration passed given to e, stamped with e's latest input tick like the
// vanilla server does for every viewer.
func (s *Session) viewGlideBoost(e world.Entity, d time.Duration) {
	ownerSession := s
	if e.H() != s.ent {
		c, ok := e.(Controllable)
		if !ok {
			return
		}
		if ownerSession, ok = sessions.Lookup(c.UUID()); !ok {
			return
		}
	}
	s.entityMutex.RLock()
	id, ok := s.entityRuntimeIDs[e.H()]
	s.entityMutex.RUnlock()
	if !ok {
		return
	}
	s.writePacket(glideBoostEffect(id, d, ownerSession.inputTick.Load()))
}

// glideBoostEffect returns the MovementEffect of a glide boost lasting d, or without end if d is negative.
func glideBoostEffect(id uint64, d time.Duration, tick uint64) *packet.MovementEffect {
	duration := int32(-1)
	if d >= 0 {
		duration = int32(d / (time.Second / 20))
	}
	return &packet.MovementEffect{EntityRuntimeID: id, Type: packet.MovementEffectTypeGlideBoost, Duration: duration, Tick: tick}
}
