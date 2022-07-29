package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
	"time"
)

// TNT represents a prime TNT entity.
type TNT struct {
	transform

	fuse int64

	c *MovementComputer
}

// NewTNT creates a new prime TNT instance.
func NewTNT(pos mgl64.Vec3, fuse time.Duration) *TNT {
	t := &TNT{
		fuse: fuse.Milliseconds() / 50,
		c: &MovementComputer{
			Gravity:           0.04,
			Drag:              0.02,
			DragBeforeGravity: true,
		},
	}
	t.transform = newTransform(t, pos)

	d := rand.Float64() * math.Pi * 2
	t.vel = mgl64.Vec3{-math.Sin(d) * 0.02, 0.1, -math.Cos(d) * 0.02}
	return t
}

// Name ...
func (t *TNT) Name() string {
	return "Primed TNT"
}

// EncodeEntity ...
func (t *TNT) EncodeEntity() string {
	return "minecraft:tnt"
}

// BBox ...
func (t *TNT) BBox() cube.BBox {
	return cube.Box(-0.49, 0, -0.49, 0.49, 0.98, 0.49)
}

// Fuse returns the remaining duration of the TNT's fuse.
func (t *TNT) Fuse() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return time.Duration(t.fuse) * time.Millisecond * 50
}

// Explode ...
func (t *TNT) Explode(explosionPos mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	t.mu.Lock()
	t.vel = t.vel.Add(t.pos.Sub(explosionPos).Normalize().Mul(impact))
	t.mu.Unlock()
}

// Tick ticks the entity, performing movement.
func (t *TNT) Tick(w *world.World, _ int64) {
	t.mu.Lock()
	m := t.c.TickMovement(t, t.pos, t.vel, 0, 0)
	t.pos, t.vel = m.pos, m.vel
	fuse := t.fuse
	t.fuse--
	t.mu.Unlock()

	m.Send()

	pos := cube.PosFromVec3(m.pos)
	if pos[1] < w.Range()[0] {
		_ = t.Close()
		return
	}

	if fuse%5 == 0 {
		for _, v := range w.Viewers(m.pos) {
			v.ViewEntityState(t)
		}
	}

	if fuse-1 <= 0 {
		_ = t.Close()

		block.ExplosionConfig{
			Size: 4,
		}.Explode(w, m.pos)
	}
}

// New creates and returns an TNT with the world.Block and position provided. It doesn't spawn the TNT by itself.
func (t *TNT) New(pos mgl64.Vec3, fuse time.Duration) world.Entity {
	return NewTNT(pos, fuse)
}

// EncodeNBT ...
func (t *TNT) EncodeNBT() map[string]any {
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(t.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(t.Velocity()),
		"Fuse":   uint8(t.Fuse().Milliseconds() / 50),
	}
}

// DecodeNBT ...
func (t *TNT) DecodeNBT(data map[string]any) any {
	tnt := NewTNT(nbtconv.MapVec3(data, "Pos"), time.Duration(nbtconv.Map[uint8](data, "Fuse"))*time.Millisecond*50)
	tnt.vel = nbtconv.MapVec3(data, "Motion")
	return tnt
}
