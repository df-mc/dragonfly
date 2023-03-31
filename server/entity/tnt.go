package entity

import (
	"math"
	"math/rand"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// TNT represents a primed TNT entity.
type TNT struct {
	transform

	uniqueID int64
	fuse     int64

	c *MovementComputer
}

// NewTNT creates a new prime TNT instance.
func NewTNT(pos mgl64.Vec3, fuse time.Duration) *TNT {
	t := &TNT{
		uniqueID: rand.Int63(),
		fuse:     fuse.Milliseconds() / 50,
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

// Type returns TNTType.
func (*TNT) Type() world.EntityType {
	return TNTType{}
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

		block.ExplosionConfig{}.Explode(w, m.pos)
	}
}

// TNTType is a world.EntityType implementation for TNT.
type TNTType struct{}

func (TNTType) EncodeEntity() string   { return "minecraft:tnt" }
func (TNTType) NetworkOffset() float64 { return 0.49 }
func (TNTType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.49, 0, -0.49, 0.49, 0.98, 0.49)
}

func (TNTType) DecodeNBT(m map[string]any) world.Entity {
	tnt := NewTNT(nbtconv.Vec3(m, "Pos"), nbtconv.TickDuration[uint8](m, "Fuse"))
	tnt.vel = nbtconv.Vec3(m, "Motion")
	if uniqueID, ok := m["UniqueID"].(int64); ok {
		tnt.uniqueID = uniqueID
	}
	return tnt
}

func (TNTType) EncodeNBT(e world.Entity) map[string]any {
	t := e.(*TNT)
	return map[string]any{
		"UniqueID": t.uniqueID,
		"Pos":      nbtconv.Vec3ToFloat32Slice(t.Position()),
		"Motion":   nbtconv.Vec3ToFloat32Slice(t.Velocity()),
		"Fuse":     uint8(t.Fuse().Milliseconds() / 50),
	}
}
