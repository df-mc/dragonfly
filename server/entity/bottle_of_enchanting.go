package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"math/rand"
)

// NewBottleOfEnchanting ...
func NewBottleOfEnchanting(opts world.EntitySpawnOpts, owner world.Entity) *world.EntityHandle {
	conf := bottleOfEnchantingConf
	conf.Owner = owner
	return opts.New(BottleOfEnchantingType{}, conf)
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
func spawnExperience(e *Ent, tx *world.Tx, target trace.Result) {
	for _, orb := range NewExperienceOrbs(target.Position(), rand.Intn(9)+3) {
		tx.AddEntity(orb)
	}
}

// BottleOfEnchantingType is a world.EntityType for BottleOfEnchanting.
type BottleOfEnchantingType struct{}

func (t BottleOfEnchantingType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

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

func (BottleOfEnchantingType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = bottleOfEnchantingConf.New()
}

func (BottleOfEnchantingType) EncodeNBT(data *world.EntityData) map[string]any {
	return nil
}
