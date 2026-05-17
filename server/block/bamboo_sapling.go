package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
)

// BambooSapling is the initial stage of bamboo growth. It appears as a small shoot
// and grows into a bamboo stalk over time.
type BambooSapling struct {
	empty
	transparent
	Age bool
}

var _ item.BoneMealAffected = BambooSapling{}

// UseOnBlock places a bamboo sapling on valid soil.
func (b BambooSapling) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, b)
	if !used {
		return false
	}
	below := pos.Side(cube.FaceDown)
	if !supportsVegetation(Bamboo{}, tx.Block(below)) && !isBambooSupport(tx.Block(below)) {
		return false
	}
	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick breaks the sapling if it loses support.
func (b BambooSapling) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	below := pos.Side(cube.FaceDown)
	if !supportsVegetation(Bamboo{}, tx.Block(below)) && !isBambooSupport(tx.Block(below)) {
		breakBlock(b, pos, tx)
		tx.PlaySound(pos.Vec3(), sound.BlockBreaking{Block: b})
	}
}

// RandomTick grows the sapling into a bamboo stalk.
// The sapling is converted to the bottom block and a new top block grows above it.
func (b BambooSapling) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if b.Age {
		return
	}
	if tx.Light(pos) < 9 {
		return
	}
	above := pos.Side(cube.FaceUp)
	if _, ok := tx.Block(above).(Air); !ok {
		return
	}
	if r.IntN(3) != 0 {
		return
	}
	// Convert sapling to the bottom bamboo block (age=0, no leaves).
	// Grow a new top block above it (age=1, small leaves).
	tx.SetBlock(pos, Bamboo{Age: false, LeafSize: bambooNoLeaves, Thick: false}, nil)
	tx.SetBlock(above, Bamboo{Age: true, LeafSize: SmallLeaves, Thick: false}, nil)
	updateBambooStalk(pos, tx)
}

// BoneMeal grows the sapling into a bamboo stalk immediately.
func (b BambooSapling) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	above := pos.Side(cube.FaceUp)
	if _, ok := tx.Block(above).(Air); !ok {
		return false
	}
	// Convert sapling to the bottom bamboo block (age=0, no leaves).
	// Grow a new top block above it (age=1, small leaves).
	tx.SetBlock(pos, Bamboo{Age: false, LeafSize: bambooNoLeaves, Thick: false}, nil)
	tx.SetBlock(above, Bamboo{Age: true, LeafSize: SmallLeaves, Thick: false}, nil)
	updateBambooStalk(pos, tx)
	return true
}

// BreakInfo ...
func (BambooSapling) BreakInfo() BreakInfo {
	b := Bamboo{}
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       oneOf(b),
		BreakHandler: func(pos cube.Pos, tx *world.Tx, u item.User) {
			tx.PlaySound(pos.Vec3(), sound.BlockBreaking{Block: b})
		},
	}
}

// HasLiquidDrops ...
func (BambooSapling) HasLiquidDrops() bool {
	return true
}

// FlammabilityInfo ...
func (BambooSapling) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// CompostChance ...
func (BambooSapling) CompostChance() float64 {
	return 0.65
}

// EncodeItem ...
func (BambooSapling) EncodeItem() (name string, meta int16) {
	return "minecraft:bamboo", 0
}

// EncodeBlock ...
func (b BambooSapling) EncodeBlock() (string, map[string]any) {
	return "minecraft:bamboo_sapling", map[string]any{"age_bit": boolByte(b.Age)}
}

// allBambooSapling returns all bamboo sapling block states.
func allBambooSapling() (blocks []world.Block) {
	for _, age := range []bool{false, true} {
		blocks = append(blocks, BambooSapling{Age: age})
	}
	return
}

// isBambooSupport checks if a block can support bamboo or bamboo sapling.
func isBambooSupport(b world.Block) bool {
	switch b.(type) {
	case Bamboo, BambooSapling:
		return true
	}
	return false
}
