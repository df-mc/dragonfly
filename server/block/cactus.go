package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// Cactus is a naturally occurring block found in deserts
type Cactus struct {
	solid

	// Age is the groth state of cactus. values from 0 to 15
	Age int
}

// UseOnBlock handles makig sure the neighbouring blocks are air.
func (c Cactus) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, c)
	if !used {
		return false
	}
	_, ok := w.Block(pos.Side(cube.FaceDown)).(Cactus)
	if !supportsVegetation(c, w.Block(pos.Side(cube.FaceDown))) && !ok {
		return false
	}
	for _, face := range cube.HorizontalFaces() {
		_, ok := w.Block(pos.Side(face)).(Air)
		if !ok {
			return false
		}
	}

	place(w, pos, c, user, ctx)
	return placed(ctx)
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
		abovePos := pos.Add(cube.Pos{0, 1})

		switch w.Block(abovePos).(type) {
		case Air:
			c.Age++
			w.SetBlock(abovePos, Cactus{Age: c.Age}, nil)
		}
	}
}

// EntityInside ...
func (c Cactus) EntityInside(_ cube.Pos, _ *world.World, e world.Entity) {
	if l, ok := e.(entity.Living); ok && !l.AttackImmune() {
		l.Hurt(0.5, damage.SourceCactus{})
	}
}

// BreakInfo ...
func (c Cactus) BreakInfo() BreakInfo {
	return newBreakInfo(0.4, alwaysHarvestable, nothingEffective, oneOf(c))
}

// EncodeItem ...
func (c Cactus) EncodeItem() (name string, meta int16) {
	return "minecraft:cactus", 0
}

// EncodeBlock ...
func (c Cactus) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:cactus", map[string]any{"age": int32(c.Age)}
}

// Model ...
func (c Cactus) Model() world.BlockModel {
	return model.Cactus{}
}

// allCactus returns all possible states of a cactus block.
func allCactus() (b []world.Block) {
	for i := 0; i < 16; i++ {
		b = append(b, Cactus{Age: i})
	}
	return
}
