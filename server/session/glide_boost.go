package session

import (
	"math"
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// glideBoost tracks the glide boost a player's own rockets give it, decided once per rocket so that every viewer
// told about the rocket sends the same MovementEffect.
type glideBoost struct {
	mu     sync.Mutex
	rocket *world.EntityHandle
	sent   bool
	effect packet.MovementEffect
	// end is the client tick the current boost ends on; math.MaxUint64 never ends.
	end uint64
}

// decide returns the boost a rocket with the lifetime passed gives from client tick, and whether it replaces the
// current boost: like the vanilla server, it does only when it lasts at least as long as what remains of it.
func (b *glideBoost) decide(rocket *world.EntityHandle, lifetime time.Duration, tick uint64) (packet.MovementEffect, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if rocket == b.rocket {
		return b.effect, b.sent
	}
	b.rocket = rocket
	duration, end := int32(-1), uint64(math.MaxUint64)
	if lifetime != 0 {
		const tickDuration = time.Second / 20
		duration = int32((max(lifetime, 0) + tickDuration - 1) / tickDuration)
		end = tick + uint64(duration)
	}
	b.sent = end >= b.end
	if b.sent {
		b.end = end
		b.effect = packet.MovementEffect{Type: packet.MovementEffectTypeGlideBoost, Duration: duration, Tick: tick}
	}
	return b.effect, b.sent
}

// viewGlideBoost shows the glide boost a rocket launched by owner gives it, as the owner's session decided it.
func (s *Session) viewGlideBoost(owner, rocket world.Entity) {
	ent, ok := rocket.(*entity.Ent)
	if !ok {
		return
	}
	f, ok := ent.Behaviour().(*entity.FireworkBehaviour)
	if !ok {
		return
	}
	id, ownerSession := uint64(selfEntityRuntimeID), s
	if owner.H() != s.ent {
		c, ok := owner.(Controllable)
		if !ok {
			return
		}
		if ownerSession, ok = sessions.Lookup(c.UUID()); !ok {
			return
		}
		s.entityMutex.RLock()
		id, ok = s.entityRuntimeIDs[owner.H()]
		s.entityMutex.RUnlock()
		if !ok {
			return
		}
	}
	effect, ok := ownerSession.glideBoost.decide(rocket.H(), f.ExistenceDuration(), ownerSession.inputTick.Load())
	if !ok {
		return
	}
	effect.EntityRuntimeID = id
	s.writePacket(&effect)
}
