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
// shown the rocket sends the same MovementEffect.
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

// viewGlideBoost shows the glide boost a newly launched rocket attached to its owner gives that owner, as the owner's
// session decided it. Every viewer of the rocket is told, as vanilla sends the boost to all players near the owner.
func (s *Session) viewGlideBoost(rocket *world.EntityHandle, f *entity.FireworkBehaviour) {
	owner := f.Owner()
	if owner == nil {
		return
	}
	id, ownerSession := uint64(selfEntityRuntimeID), s
	if owner != s.ent {
		var ok bool
		if ownerSession, ok = sessions.Lookup(owner.UUID()); !ok {
			return
		}
		s.entityMutex.RLock()
		id, ok = s.entityRuntimeIDs[owner]
		s.entityMutex.RUnlock()
		if !ok {
			return
		}
	}
	effect, ok := ownerSession.glideBoost.decide(rocket, f.ExistenceDuration(), ownerSession.inputTick.Load())
	if !ok {
		return
	}
	effect.EntityRuntimeID = id
	s.writePacket(&effect)
}
