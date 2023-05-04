package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// NewFirework creates a firework entity. Firework is an item (and entity) used
// for creating decorative explosions, boosting when flying with elytra, and
// loading into a crossbow as ammunition.
func NewFirework(pos mgl64.Vec3, rot cube.Rotation, firework item.Firework) *Ent {
	return NewFireworkAttached(pos, rot, firework, nil, false)
}

// NewFireworkAttached creates a firework entity with an owner that the firework
// may be attached to.
func NewFireworkAttached(pos mgl64.Vec3, rot cube.Rotation, firework item.Firework, owner world.Entity, attached bool) *Ent {
	e := Config{Behaviour: FireworkBehaviourConfig{
		ExistenceDuration:          firework.RandomisedDuration(),
		SidewaysVelocityMultiplier: 1.15,
		UpwardsAcceleration:        0.04,
		Attached:                   attached,
	}.New(firework, owner)}.New(FireworkType{}, pos)
	e.rot = rot
	return e
}

// FireworkType is a world.EntityType implementation for Firework.
type FireworkType struct{}

func (FireworkType) EncodeEntity() string        { return "minecraft:fireworks_rocket" }
func (FireworkType) BBox(world.Entity) cube.BBox { return cube.BBox{} }

func (FireworkType) DecodeNBT(m map[string]any) world.Entity {
	f := NewFirework(
		nbtconv.Vec3(m, "Pos"),
		nbtconv.Rotation(m),
		nbtconv.MapItem(m, "Item").Item().(item.Firework),
	)
	f.vel = nbtconv.Vec3(m, "Motion")
	return f
}

func (FireworkType) EncodeNBT(e world.Entity) map[string]any {
	f := e.(*Ent)
	yaw, pitch := f.Rotation().Elem()
	return map[string]any{
		"Item":   nbtconv.WriteItem(item.NewStack(f.Behaviour().(*FireworkBehaviour).Firework(), 1), true),
		"Pos":    nbtconv.Vec3ToFloat32Slice(f.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(f.Velocity()),
		"Yaw":    float32(yaw),
		"Pitch":  float32(pitch),
	}
}
