package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"slices"
	"time"
)

// orbSplitSizes contains split sizes used for dropping experience orbs.
var orbSplitSizes = []int{2477, 1237, 617, 307, 149, 73, 37, 17, 7, 3, 1}

// NewExperienceOrbs takes in a position and an amount and automatically splits the amount into multiple orbs, returning
// a slice of the created orbs.
func NewExperienceOrbs(pos mgl64.Vec3, amount int) (orbs []*Ent) {
	for amount > 0 {
		size := orbSplitSizes[slices.IndexFunc(orbSplitSizes, func(value int) bool {
			return amount >= value
		})]

		orbs = append(orbs, NewExperienceOrb(pos, size))
		amount -= size
	}
	return
}

// NewExperienceOrb creates a new experience orb and returns it.
func NewExperienceOrb(pos mgl64.Vec3, xp int) *Ent {
	conf := experienceOrbConf
	conf.Experience = xp
	return Config{Behaviour: conf.New()}.New(ExperienceOrbType{}, pos)
}

var experienceOrbConf = ExperienceOrbBehaviourConfig{
	Gravity: 0.04,
	Drag:    0.02,
}

// ExperienceOrbType is a world.EntityType implementation for ExperienceOrb.
type ExperienceOrbType struct{}

func (ExperienceOrbType) EncodeEntity() string { return "minecraft:xp_orb" }
func (ExperienceOrbType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (ExperienceOrbType) DecodeNBT(m map[string]any) world.Entity {
	o := NewExperienceOrb(nbtconv.Vec3(m, "Pos"), int(nbtconv.Int32(m, "Value")))
	o.vel = nbtconv.Vec3(m, "Motion")
	o.age = time.Duration(nbtconv.Int16(m, "Age")) * (time.Second / 20)
	return o
}

func (ExperienceOrbType) EncodeNBT(e world.Entity) map[string]any {
	orb := e.(*Ent)
	return map[string]any{
		"Age":    int16(orb.Age() / (time.Second * 20)),
		"Value":  int32(orb.Behaviour().(*ExperienceOrbBehaviour).Experience()),
		"Pos":    nbtconv.Vec3ToFloat32Slice(orb.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(orb.Velocity()),
	}
}
