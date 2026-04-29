package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Sapling is a non-solid transparent plant block that ages through age_bit and can grow into a tree.
type Sapling struct {
	replaceable
	transparent
	empty

	// Type is the sapling species.
	Type SaplingType
	// Aged is the Bedrock age_bit growth stage.
	Aged bool
}

// BreakInfo ...
func (s Sapling) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(s))
}

// NeighbourUpdateTick ...
func (s Sapling) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsVegetation(s, tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(s, pos, tx)
	}
}

// HasLiquidDrops ...
func (Sapling) HasLiquidDrops() bool {
	return true
}

// BoneMeal attempts to affect the sapling using bone meal.
func (s Sapling) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	if !saplingGrowthBaseValid(pos, tx) {
		return false
	}
	if rand.Float64() >= 0.45 {
		return true
	}
	if !s.Aged {
		s.Aged = true
		tx.SetBlock(pos, s, nil)
		return true
	}
	_ = growSaplingTree(pos, tx, s.Type, rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())))
	return true
}

// RandomTick ...
func (s Sapling) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if saplingShouldUproot(pos, tx) {
		breakBlock(s, pos, tx)
		return
	}
	if !saplingGrowthAllowed(pos, tx) {
		return
	}
	if !saplingGrowthBaseValid(pos, tx) {
		return
	}
	if !s.Aged {
		s.Aged = true
		tx.SetBlock(pos, s, nil)
		return
	}
	_ = growSaplingTree(pos, tx, s.Type, r)
}

// UseOnBlock ...
func (s Sapling) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, s)
	if !used || !supportsVegetation(s, tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	place(tx, pos, Sapling{Type: s.Type}, user, ctx)
	return placed(ctx)
}

// FuelInfo ...
func (Sapling) FuelInfo() item.FuelInfo {
	return newFuelInfo(5 * time.Second)
}

// CompostChance ...
func (Sapling) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (s Sapling) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Type.String(), 0
}

// EncodeBlock ...
func (s Sapling) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + s.Type.String(), map[string]any{"age_bit": s.Aged}
}

// allSaplings ...
func allSaplings() (saplings []world.Block) {
	for _, t := range SaplingTypes() {
		saplings = append(saplings, Sapling{Type: t})
		saplings = append(saplings, Sapling{Type: t, Aged: true})
	}
	return
}
