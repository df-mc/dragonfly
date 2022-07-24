package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// Firework is an item (and entity) used for creating decorative explosions, boosting when flying with elytra, and
// loading into a crossbow as ammunition.
type Firework struct {
	transform

	yaw, pitch float64
	firework   item.Firework

	c *MovementComputer

	ticks int
	close bool
}

// NewFirework ...
func NewFirework(pos mgl64.Vec3, yaw, pitch float64, firework item.Firework) *Firework {
	f := &Firework{
		yaw:      yaw,
		pitch:    pitch,
		firework: firework,
		c:        &MovementComputer{},
		ticks:    int(firework.RandomizedDuration().Milliseconds() / 50),
	}
	f.transform = newTransform(f, pos)
	f.vel = mgl64.Vec3{rand.Float64() * 0.001, 0.05, rand.Float64() * 0.001}
	return f
}

// Name ...
func (f *Firework) Name() string {
	return "Firework Rocket"
}

// EncodeEntity ...
func (f *Firework) EncodeEntity() string {
	return "minecraft:fireworks_rocket"
}

// BBox ...
func (f *Firework) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Firework returns the underlying item.Firework of the Firework.
func (f *Firework) Firework() item.Firework {
	return f.firework
}

// Rotation ...
func (f *Firework) Rotation() (float64, float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.yaw, f.pitch
}

// Tick ...
func (f *Firework) Tick(w *world.World, current int64) {
	if f.close {
		_ = f.Close()
		return
	}

	f.mu.Lock()
	f.vel[0] *= 1.15
	f.vel[1] += 0.04
	f.vel[2] *= 1.15
	m := f.c.TickMovement(f, f.pos, f.vel, f.yaw, f.pitch)
	f.pos, f.vel, f.yaw, f.pitch = m.pos, m.vel, m.yaw, m.pitch
	f.mu.Unlock()

	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		f.close = true
		return
	}

	f.ticks--
	if f.ticks < 0 {
		for _, v := range w.Viewers(m.pos) {
			v.ViewEntityAction(f, FireworkParticleAction{})
		}
		for _, explosion := range f.Firework().Explosions {
			if explosion.Shape == item.FireworkShapeHugeSphere() {
				w.PlaySound(m.pos, sound.FireworkHugeBlast{})
			} else {
				w.PlaySound(m.pos, sound.FireworkBlast{})
			}
			if explosion.Twinkle {
				w.PlaySound(m.pos, sound.FireworkTwinkle{})
			}
		}
		f.close = true
	}
}

// New creates an firework with the position, velocity, yaw, and pitch provided. It doesn't spawn the firework,
// only returns it.
func (f *Firework) New(pos mgl64.Vec3, yaw, pitch float64, firework item.Firework) world.Entity {
	return NewFirework(pos, yaw, pitch, firework)
}

// DecodeNBT decodes the properties in a map to a Firework and returns a new Firework entity.
func (f *Firework) DecodeNBT(data map[string]any) any {
	firework := NewFirework(
		nbtconv.MapVec3(data, "Pos"),
		float64(nbtconv.Map[float32](data, "Pitch")),
		float64(nbtconv.Map[float32](data, "Yaw")),
		nbtconv.MapItem(data, "Item").Item().(item.Firework),
	)
	firework.vel = nbtconv.MapVec3(data, "Motion")
	return firework
}

// EncodeNBT encodes the Firework entity's properties as a map and returns it.
func (f *Firework) EncodeNBT() map[string]any {
	yaw, pitch := f.Rotation()
	return map[string]any{
		"Item":   nbtconv.WriteItem(item.NewStack(f.Firework(), 1), true),
		"Pos":    nbtconv.Vec3ToFloat32Slice(f.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(f.Velocity()),
		"Yaw":    yaw,
		"Pitch":  pitch,
	}
}
