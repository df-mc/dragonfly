package entity

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// cushionBehaviour implements the behaviour of a cushion.
type cushionBehaviour struct {
	colour item.Colour
	rider  *world.EntityHandle

	// survivalDelay is the number of ticks until the next support check. It starts at 0, so the first tick checks.
	survivalDelay int
	exploded      bool
	removed       bool
}

// Variant ...
func (b *cushionBehaviour) Variant() int32 {
	return int32(15 - b.colour.Uint8())
}

// Gravityless ...
func (b *cushionBehaviour) Gravityless() bool {
	return true
}

// ServerAuthOnlyDismount ...
func (b *cushionBehaviour) ServerAuthOnlyDismount() bool {
	return true
}

// Tick breaks the cushion in lava or a lit campfire, and checks its support every 81-121 ticks like vanilla.
func (b *cushionBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	b.checkRider(e, tx)
	if b.exploded {
		b.destroy(e, tx, true)
		return nil
	}
	if b.hurtByBlocks(e, tx) {
		return nil
	}
	if b.survivalDelay > 0 {
		b.survivalDelay--
		return nil
	}
	b.survivalDelay = 80 + rand.IntN(41)
	if !cushionCanSurvive(e.Position(), tx) {
		b.destroy(e, tx, true)
	}
	return nil
}

// checkRider frees the seat if the rider is gone or no longer sitting on the cushion.
func (b *cushionBehaviour) checkRider(e *Ent, tx *world.Tx) {
	if b.rider == nil {
		return
	}
	ent, ok := b.rider.Entity(tx)
	if !ok {
		b.rider = nil
		return
	}
	if r, ok := ent.(Rider); !ok || r.RidingEntityHandle() != e.H() {
		b.rider = nil
	}
}

// hurtByBlocks hurts the cushion in lava or a lit campfire. Unlike Java, fire and magma do not hurt cushions.
func (b *cushionBehaviour) hurtByBlocks(e *Ent, tx *world.Tx) bool {
	box := CushionType.BBox(e).Translate(e.Position()).Grow(-1e-4)
	for pos := range cube.Range3D(cube.PosFromVec3(box.Min()), cube.PosFromVec3(box.Max())) {
		if l, ok := tx.Liquid(pos); ok {
			if _, lava := l.(block.Lava); lava {
				b.Hurt(e, 4, block.LavaDamageSource{})
				return true
			}
		}
		if c, ok := tx.Block(pos).(block.Campfire); ok && !c.Extinguished {
			b.Hurt(e, c.Type.Damage(), block.FireDamageSource{})
			return true
		}
	}
	return false
}

// Hurt breaks the cushion on any damage. Adventure players and ender pearls cannot break it, and creative
// players break it without a drop.
func (b *cushionBehaviour) Hurt(e *Ent, damage float64, src world.DamageSource) (float64, bool) {
	damage = max(damage, 0)
	var attacker world.Entity
	switch s := src.(type) {
	case VoidDamageSource:
		b.remove(e, e.tx)
		return damage, true
	case AttackDamageSource:
		attacker = s.Attacker
	case ProjectileDamageSource:
		if s.Projectile != nil && s.Projectile.H().Type() == EnderPearlType {
			return 0, false
		}
		attacker = s.Owner
	}
	drop := true
	if g, ok := attacker.(interface{ GameMode() world.GameMode }); ok {
		if !g.GameMode().AllowsEditing() {
			return 0, false
		}
		drop = !g.GameMode().CreativeInventory()
	}
	b.destroy(e, e.tx, drop)
	return damage, true
}

// Explode breaks the cushion on its next tick, even if the blast is fully blocked. The rider stays seated during the
// explosion, so like vanilla it is not pushed by it.
func (b *cushionBehaviour) Explode(*Ent, world.ExplosionSource, float64) {
	b.exploded = true
}

// destroy breaks the cushion into wool particles, optionally dropping it with its name.
func (b *cushionBehaviour) destroy(e *Ent, tx *world.Tx, drop bool) {
	if b.removed {
		return
	}
	pos := e.Position()
	centre := pos.Add(mgl64.Vec3{0, cushionHeight / 2, 0})
	tx.AddParticle(centre, particle.BlockBreakNoSound{Block: block.Wool{Colour: b.colour}})
	tx.PlaySound(centre, sound.CushionBreak{})
	if drop {
		stack := item.NewStack(item.Cushion{Colour: b.colour}, 1)
		if name := e.NameTag(); name != "" {
			stack = stack.WithCustomName(name)
		}
		vel := mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1}
		tx.AddEntity(NewItem(world.EntitySpawnOpts{Position: pos, Velocity: vel}, stack))
	}
	b.remove(e, tx)
}

// remove gets the rider off and removes the cushion from the world.
func (b *cushionBehaviour) remove(e *Ent, tx *world.Tx) {
	if b.removed {
		return
	}
	b.removed = true
	if b.rider != nil {
		if ent, ok := b.rider.Entity(tx); ok {
			if r, ok := ent.(Rider); ok && r.RidingEntityHandle() == e.H() {
				r.DismountEntity(tx)
			}
		}
		b.rider = nil
	}
	_ = e.Close()
}
