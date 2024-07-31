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

// NewTNT creates a new primed TNT entity.
func NewTNT(pos mgl64.Vec3, fuse time.Duration, igniter world.Entity) *Ent {
	config := tntConf
	config.ExistenceDuration = fuse
	ent := Config{Behaviour: config.New()}.New(TNTType{igniter: igniter}, pos)

	angle := rand.Float64() * math.Pi * 2
	ent.vel = mgl64.Vec3{-math.Sin(angle) * 0.02, 0.1, -math.Cos(angle) * 0.02}
	return ent
}

var tntConf = PassiveBehaviourConfig{
	Gravity: 0.04,
	Drag:    0.02,
	Expire:  explodeTNT,
}

// explodeTNT creates an explosion at the position of e.
func explodeTNT(e *Ent) {
	var config block.ExplosionConfig
	config.Explode(e.World(), e.Position())
}

// TNTType is a world.EntityType implementation for TNT.
type TNTType struct {
	igniter world.Entity
}

// Igniter returns the entity that ignited the TNT.
// It is nil if ignited by a world source like fire.
func (t TNTType) Igniter() world.Entity { return t.igniter }

func (TNTType) EncodeEntity() string   { return "minecraft:tnt" }
func (TNTType) NetworkOffset() float64 { return 0.49 }
func (TNTType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.49, 0, -0.49, 0.49, 0.98, 0.49)
}

func (t TNTType) DecodeNBT(m map[string]any) world.Entity {
	tnt := NewTNT(nbtconv.Vec3(m, "Pos"), nbtconv.TickDuration[uint8](m, "Fuse"), t.igniter)
	tnt.vel = nbtconv.Vec3(m, "Motion")
	return tnt
}

func (TNTType) EncodeNBT(e world.Entity) map[string]any {
	t := e.(*Ent)
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(t.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(t.Velocity()),
		"Fuse":   uint8(t.Behaviour().(*PassiveBehaviour).Fuse().Milliseconds() / 50),
	}
}
