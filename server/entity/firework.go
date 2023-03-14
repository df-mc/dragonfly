package entity

import (
	"math"
	"math/rand"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Firework is an item (and entity) used for creating decorative explosions, boosting when flying with elytra, and
// loading into a crossbow as ammunition.
type Firework struct {
	transform

	uniqueID   int64
	yaw, pitch float64
	firework   item.Firework

	owner world.Entity

	c *MovementComputer

	attached bool

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
		ticks:    int(firework.RandomisedDuration().Milliseconds() / 50),
	}
	f.transform = newTransform(f, pos)
	f.vel = mgl64.Vec3{rand.Float64() * 0.001, 0.05, rand.Float64() * 0.001}
	return f
}

// Type returns FireworkType.
func (f *Firework) Type() world.EntityType {
	return FireworkType{}
}

// Firework returns the underlying item.Firework of the Firework.
func (f *Firework) Firework() item.Firework {
	return f.firework
}

// Rotation ...
func (f *Firework) Rotation() cube.Rotation {
	f.mu.Lock()
	defer f.mu.Unlock()
	return cube.Rotation{f.yaw, f.pitch}
}

// Tick ...
func (f *Firework) Tick(w *world.World, current int64) {
	if f.close {
		_ = f.Close()
		return
	}

	f.mu.Lock()
	if f.attached {
		if o, ok := f.owner.(interface {
			Velocity() mgl64.Vec3
		}); ok {
			vel := o.Velocity()
			dV := f.owner.Rotation().Vec3()

			// The client will propel itself to match the firework's velocity since we set the appropriate metadata.
			f.pos = f.owner.Position()
			f.vel.Add(vel.Add(dV.Mul(0.1).Add(dV.Mul(1.5).Sub(vel).Mul(0.5))))
		}
	} else {
		f.vel[0] *= 1.15
		f.vel[1] += 0.04
		f.vel[2] *= 1.15
	}
	m := f.c.TickMovement(f, f.pos, f.vel, f.yaw, f.pitch)
	f.pos, f.vel = m.pos, m.vel
	f.mu.Unlock()

	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		f.close = true
		return
	}

	f.ticks--
	if f.ticks >= 0 {
		return
	}

	explosions := f.Firework().Explosions
	for _, v := range w.Viewers(m.pos) {
		v.ViewEntityAction(f, FireworkExplosionAction{})
	}
	for _, explosion := range explosions {
		if explosion.Shape == item.FireworkShapeHugeSphere() {
			w.PlaySound(m.pos, sound.FireworkHugeBlast{})
		} else {
			w.PlaySound(m.pos, sound.FireworkBlast{})
		}
		if explosion.Twinkle {
			w.PlaySound(m.pos, sound.FireworkTwinkle{})
		}
	}

	if len(explosions) > 0 {
		force := float64(len(explosions)*2) + 5.0
		for _, e := range w.EntitiesWithin(f.Type().BBox(f).Translate(m.pos).Grow(5.25), func(e world.Entity) bool {
			l, living := e.(Living)
			return !living || l.AttackImmune()
		}) {
			pos := e.Position()
			dist := m.pos.Sub(pos).Len()
			if dist > 5.0 {
				// The maximum distance allowed is 5.0 blocks.
				continue
			}
			if _, ok := trace.Perform(m.pos, pos, w, e.Type().BBox(e).Grow(0.3), func(world.Entity) bool {
				return true
			}); ok {
				dmg := force * math.Sqrt((5.0-dist)/5.0)
				e.(Living).Hurt(dmg, ProjectileDamageSource{Owner: f.Owner(), Projectile: f})
			}
		}
	}

	f.close = true
}

// Attached returns true if the firework is currently attached to the owner. This is mainly the case with gliding.
func (f *Firework) Attached() bool {
	return f.attached
}

// Owner ...
func (f *Firework) Owner() world.Entity {
	return f.owner
}

// FireworkType is a world.EntityType implementation for Firework.
type FireworkType struct{}

func (FireworkType) EncodeEntity() string        { return "minecraft:fireworks_rocket" }
func (FireworkType) BBox(world.Entity) cube.BBox { return cube.BBox{} }

func (FireworkType) DecodeNBT(m map[string]any) world.Entity {
	f := NewFirework(
		nbtconv.Vec3(m, "Pos"),
		float64(nbtconv.Float32(m, "Pitch")),
		float64(nbtconv.Float32(m, "Yaw")),
		nbtconv.MapItem(m, "Item").Item().(item.Firework),
	)
	f.vel = nbtconv.Vec3(m, "Motion")
	if uniqueID, ok := m["UniqueID"].(int64); ok {
		f.uniqueID = uniqueID
	}
	return f
}

func (FireworkType) EncodeNBT(e world.Entity) map[string]any {
	f := e.(*Firework)
	yaw, pitch := f.Rotation().Elem()
	return map[string]any{
		"UniqueID": f.uniqueID,
		"Item":     nbtconv.WriteItem(item.NewStack(f.Firework(), 1), true),
		"Pos":      nbtconv.Vec3ToFloat32Slice(f.Position()),
		"Motion":   nbtconv.Vec3ToFloat32Slice(f.Velocity()),
		"Yaw":      float32(yaw),
		"Pitch":    float32(pitch),
	}
}
