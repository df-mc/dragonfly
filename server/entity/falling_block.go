package entity

import (
	"fmt"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
)

// FallingBlock is the entity form of a block that appears when a gravity-affected block loses its support.
type FallingBlock struct {
	transform

	block world.Block

	fallDistance atomic.Float64

	damagePerDistance float64
	damageMax         float64
	damages           bool

	c *MovementComputer
}

// NewFallingBlock creates a new FallingBlock entity.
func NewFallingBlock(block world.Block, pos mgl64.Vec3) *FallingBlock {
	b := &FallingBlock{
		block: block,
		c: &MovementComputer{
			Gravity:           0.04,
			Drag:              0.02,
			DragBeforeGravity: true,
		},
	}
	b.transform = newTransform(b, pos)
	return b
}

// Name ...
func (f *FallingBlock) Name() string {
	return fmt.Sprintf("%T", f.block)
}

// EncodeEntity ...
func (f *FallingBlock) EncodeEntity() string {
	return "minecraft:falling_block"
}

// BBox ...
func (f *FallingBlock) BBox() cube.BBox {
	return cube.Box(-0.49, 0, -0.49, 0.49, 0.98, 0.49)
}

// Block ...
func (f *FallingBlock) Block() world.Block {
	return f.block
}

// damageable ...
type damageable interface {
	Damage() world.Block
}

// landable ...
type landable interface {
	Landed(w *world.World, pos cube.Pos)
}

// SetDamage ...
func (f *FallingBlock) SetDamage(damagePerDistance, damageMax float64) {
	f.damagePerDistance = damagePerDistance
	f.damageMax = damageMax
	f.damages = true
}

// Tick ...
func (f *FallingBlock) Tick(w *world.World, _ int64) {
	f.mu.Lock()
	m := f.c.TickMovement(f, f.pos, f.vel, 0, 0)
	f.pos, f.vel = m.pos, m.vel
	f.mu.Unlock()

	m.Send()

	distThisTick := f.vel.Y()
	if distThisTick < f.fallDistance.Load() {
		f.fallDistance.Sub(distThisTick)
	} else {
		f.fallDistance.Store(0)
	}

	pos := cube.PosFromVec3(m.pos)
	if pos[1] < w.Range()[0] {
		_ = f.Close()
	}

	if a, ok := f.block.(Solidifiable); (ok && a.Solidifies(pos, w)) || f.c.OnGround() {
		if f.damages {
			fallDist := math.Ceil(f.fallDistance.Load() - 1.0)
			if fallDist > 0 {
				force := math.Min(math.Floor(fallDist*f.damagePerDistance), f.damageMax)

				for _, e := range w.EntitiesWithin(f.BBox().Translate(m.pos).Grow(0.05), f.ignores) {
					l := e.(Living)
					l.Hurt(force, damage.SourceDamagingBlock{Block: f.block})
				}

				if d, ok := f.block.(damageable); ok && force > 0.0 && rand.Float64() < 0.05+fallDist*0.05 {
					f.block = d.Damage()
				}
			}
		}

		if l, ok := f.block.(landable); ok {
			l.Landed(w, pos)
		}

		b := w.Block(pos)
		if r, ok := b.(replaceable); ok && r.ReplaceableBy(f.block) {
			w.SetBlock(pos, f.block, nil)
		} else {
			if i, ok := f.block.(world.Item); ok {
				w.AddEntity(NewItem(item.NewStack(i, 1), pos.Vec3Middle()))
			}
		}

		_ = f.Close()
	}
}

// TODO: Contain values for damaging in NBT.

// DecodeNBT decodes the relevant data from the entity NBT passed and returns a new FallingBlock entity.
func (f *FallingBlock) DecodeNBT(data map[string]any) any {
	b := nbtconv.MapBlock(data, "FallingBlock")
	if b == nil {
		return nil
	}
	n := NewFallingBlock(b, nbtconv.MapVec3(data, "Pos"))
	n.SetVelocity(nbtconv.MapVec3(data, "Motion"))
	return n
}

// EncodeNBT encodes the FallingBlock entity to a map that can be encoded for NBT.
func (f *FallingBlock) EncodeNBT() map[string]any {
	return map[string]any{
		"UniqueID":     -rand.Int63(),
		"Pos":          nbtconv.Vec3ToFloat32Slice(f.Position()),
		"Motion":       nbtconv.Vec3ToFloat32Slice(f.Velocity()),
		"FallingBlock": nbtconv.WriteBlock(f.block),
	}
}

// ignores returns whether the FallingBlock should ignore collision with the entity passed.
func (f *FallingBlock) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == f
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
