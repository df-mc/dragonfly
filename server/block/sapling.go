package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Sapling is a non-solid block that grows into the tree of its type.
type Sapling struct {
	empty
	transparent

	// Type is the type of the sapling.
	Type SaplingType
	// Aged specifies if the sapling has passed its first growth stage.
	// A sapling grows into a tree only once it has.
	Aged bool
}

var (
	_ item.BoneMealAffected = Sapling{}
	_ Flammable             = Sapling{}
)

// BoneMeal ...
func (s Sapling) BoneMeal(pos cube.Pos, tx *world.Tx) item.BoneMealResult {
	if !s.growable(pos, tx) {
		return item.BoneMealResultNone
	}
	if rand.Float64() < 0.45 && !s.grow(pos, tx) {
		return item.BoneMealResultNone
	}
	return item.BoneMealResultSmall
}

// FlammabilityInfo ...
func (s Sapling) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 60, true)
}

// NeighbourUpdateTick ...
func (s Sapling) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsVegetation(s, tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(s, pos, tx)
	}
}

// RandomTick ...
func (s Sapling) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if tx.Light(pos) >= 8 && r.IntN(7) == 0 {
		_ = s.grow(pos, tx)
	}
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

// BreakInfo ...
func (s Sapling) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(Sapling{Type: s.Type}))
}

// FuelInfo ...
func (Sapling) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 5)
}

// CompostChance ...
func (Sapling) CompostChance() float64 {
	return 0.3
}

// HasLiquidDrops ...
func (Sapling) HasLiquidDrops() bool {
	return true
}

// EncodeItem ...
func (s Sapling) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Type.String(), 0
}

// EncodeBlock ...
func (s Sapling) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + s.Type.String(), map[string]any{"age_bit": s.Aged}
}

// growable returns whether the sapling can grow into a tree at all. A dark oak and a pale oak grow from a two by
// two square of saplings and from nothing else.
func (s Sapling) growable(pos cube.Pos, tx *world.Tx) bool {
	switch s.Type {
	case DarkOakSapling(), PaleOakSapling():
		_, ok := saplingSquare(pos, tx, s.Type)
		return ok
	}
	return true
}

// grow passes the sapling to its next growth stage, growing a tree if it was already aged. It reports whether the
// sapling changed at all: a tree that does not fit leaves the sapling as it was, rather than falling back to its
// first stage.
func (s Sapling) grow(pos cube.Pos, tx *world.Tx) bool {
	if !s.Aged {
		s.Aged = true
		tx.SetBlock(pos, s, nil)
		return true
	}
	return growTree(s.Type, pos, tx)
}

// allSaplings returns a list of all sapling states.
func allSaplings() (saplings []world.Block) {
	for _, t := range SaplingTypes() {
		saplings = append(saplings, Sapling{Type: t}, Sapling{Type: t, Aged: true})
	}
	return
}
