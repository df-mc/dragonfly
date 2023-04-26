package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
)

// FallingBlockBehaviourConfig holds optional parameters for
// FallingBlockBehaviour.
type FallingBlockBehaviourConfig struct {
	// Gravity is the amount of Y velocity subtracted every tick.
	Gravity float64
	// Drag is used to reduce all axes of the velocity every tick. Velocity is
	// multiplied with (1-Drag) every tick.
	Drag float64
}

// New creates a FallingBlockBehaviour using the optional parameters in conf and
// a block type.
func (conf FallingBlockBehaviourConfig) New(b world.Block) *FallingBlockBehaviour {
	behaviour := &FallingBlockBehaviour{block: b}
	behaviour.passive = PassiveBehaviourConfig{
		Gravity: conf.Gravity,
		Drag:    conf.Drag,
		Tick:    behaviour.tick,
	}.New()
	return behaviour
}

// FallingBlockBehaviour implements the behaviour for falling block entities.
type FallingBlockBehaviour struct {
	passive *PassiveBehaviour
	block   world.Block
}

// Block returns the world.Block of the entity.
func (f *FallingBlockBehaviour) Block() world.Block {
	return f.block
}

// Tick implements the movement and solidification behaviour of falling blocks.
func (f *FallingBlockBehaviour) Tick(e *Ent) *Movement {
	return f.passive.Tick(e)
}

// tick checks if the falling block should solidify.
func (f *FallingBlockBehaviour) tick(e *Ent) {
	pos := e.Position()
	bpos, w := cube.PosFromVec3(pos), e.World()
	if a, ok := f.block.(Solidifiable); (ok && a.Solidifies(bpos, w)) || f.passive.mc.OnGround() {
		f.solidify(e, pos, w)
	}
}

// solidify attempts to solidify the falling block at the position passed. It
// also deals damage to any entities standing at that position. If the block at
// the position could not be replaced by the falling block, the block will drop
// as an item.
func (f *FallingBlockBehaviour) solidify(e *Ent, pos mgl64.Vec3, w *world.World) {
	bpos := cube.PosFromVec3(pos)

	if d, ok := f.block.(damager); ok {
		f.damageEntities(e, d, pos, w)
	}
	if l, ok := f.block.(landable); ok {
		l.Landed(w, bpos)
	}
	f.passive.close = true

	if r, ok := w.Block(bpos).(replaceable); ok && r.ReplaceableBy(f.block) {
		w.SetBlock(bpos, f.block, nil)
	} else if i, ok := f.block.(world.Item); ok {
		w.AddEntity(NewItem(item.NewStack(i, 1), bpos.Vec3Middle()))
	}
}

// damageEntities attempts to damage any entities standing below the falling
// block. This functionality is used by falling anvils.
func (f *FallingBlockBehaviour) damageEntities(e *Ent, d damager, pos mgl64.Vec3, w *world.World) {
	damagePerBlock, maxDamage := d.Damage()
	dist := math.Ceil(f.passive.fallDistance - 1.0)
	if dist <= 0 {
		return
	}
	dmg := math.Min(math.Floor(dist*damagePerBlock), maxDamage)
	targets := w.EntitiesWithin(e.Type().BBox(e).Translate(pos).Grow(0.05), func(entity world.Entity) bool {
		_, ok := entity.(Living)
		return !ok || entity == e
	})
	src := block.DamageSource{Block: f.block}

	for _, e := range targets {
		e.(Living).Hurt(dmg, src)
	}
	if b, ok := f.block.(breakable); ok && dmg > 0.0 && rand.Float64() < (dist+1)*0.05 {
		f.block = b.Break()
	}
}

// Solidifiable represents a block that can solidify by specific adjacent blocks. An example is concrete
// powder, which can turn into concrete by touching water.
type Solidifiable interface {
	// Solidifies returns whether the falling block can solidify at the position it is currently in. If so,
	// the block will immediately stop falling.
	Solidifies(pos cube.Pos, w *world.World) bool
}

type replaceable interface {
	ReplaceableBy(b world.Block) bool
}

// damager ...
type damager interface {
	Damage() (damagePerBlock, maxDamage float64)
}

// breakable ...
type breakable interface {
	Break() world.Block
}

// landable ...
type landable interface {
	Landed(w *world.World, pos cube.Pos)
}
