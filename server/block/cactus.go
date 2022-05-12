package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"math/rand"
)

// A Cactus is a naturally occurring block found in deserts
type Cactus struct {
	solid
	snare

	// Age is the groth state of cactus. values from 0 to 15 maybe?
	Age int
}

// FlammabilityInfo ...
func (c Cactus) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(5, 5, true)
}

// NeighbourUpdateTick ...
func (c Cactus) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	for _, face := range cube.HorizontalFaces() {
		_, ok := w.Block(pos.Side(face)).(Air)
		if !ok {
			w.SetBlock(pos, nil, nil)
			w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: c})
			return
		}
	}
	if w.Block(pos.Side(cube.FaceDown)) != c && !supportsVegetation(c, w.Block(pos.Side(cube.FaceDown))) {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: c})
	}
}

// RandomTick ...
func (c Cactus) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	if c.Age < 15 && r.Float64() < 0.01 {
		c.Age++
		w.SetBlock(pos, Cactus{Age: c.Age + 1}, nil)
	}
}

// BreakInfo ...
func (c Cactus) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(c))
}

// EncodeItem ...
func (c Cactus) EncodeItem() (name string, meta int16) {
	return "minecraft:cactus", 0
}

// EncodeBlock ...
func (c Cactus) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:cactus", map[string]any{"age": int32(c.Age)}
}

// allCactus returns all possible states of a cactus block.
func allCactus() (b []world.Block) {
	for i := 0; i < 16; i++ {
		b = append(b, Cactus{Age: i})
	}
	return
}
