package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// NewBottleOfEnchanting ...
func NewBottleOfEnchanting(pos mgl64.Vec3, owner world.Entity) *Ent {
	return Config{Behaviour: bottleOfEnchantingConf.New(owner)}.New(BottleOfEnchantingType{}, pos)
}

var bottleOfEnchantingConf = ProjectileBehaviourConfig{
	Gravity:  0.07,
	Drag:     0.01,
	Particle: particle.Splash{},
	Sound:    sound.GlassBreak{},
	Hit:      spawnExperience,
	Damage:   -1,
}

// spawnExperience spawns experience orbs with a value of 3-11 at the target of
// a trace.Result.
func spawnExperience(e *Ent, target trace.Result) {
	for _, orb := range NewExperienceOrbs(target.Position(), rand.Intn(9)+3) {
		orb.SetVelocity(mgl64.Vec3{(rand.Float64()*0.2 - 0.1) * 2, rand.Float64() * 0.4, (rand.Float64()*0.2 - 0.1) * 2})
		e.World().AddEntity(orb)
	}
}

// BottleOfEnchantingType is a world.EntityType for BottleOfEnchanting.
type BottleOfEnchantingType struct{}

// Glint returns true if the bottle should render with glint. It always returns
// true for bottles of enchanting.
func (BottleOfEnchantingType) Glint() bool {
	return true
}
func (BottleOfEnchantingType) EncodeEntity() string {
	return "minecraft:xp_bottle"
}
func (BottleOfEnchantingType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (BottleOfEnchantingType) DecodeNBT(m map[string]any) world.Entity {
	b := NewBottleOfEnchanting(nbtconv.Vec3(m, "Pos"), nil)
	b.vel = nbtconv.Vec3(m, "Motion")
	return b
}

func (BottleOfEnchantingType) EncodeNBT(e world.Entity) map[string]any {
	b := e.(*Ent)
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(b.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(b.Velocity()),
	}
}
