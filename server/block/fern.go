package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// Fern is a transparent plant block which can be used to obtain seeds and as decoration.
type Fern struct {
	replaceable
	transparent
	empty
}

// FlammabilityInfo ...
func (g Fern) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// BreakInfo ...
func (g Fern) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if t.ToolType() == item.TypeShears || hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(g, 1)}
		}
		if rand.Float32() > 0.57 {
			return []item.Stack{item.NewStack(WheatSeeds{}, 1)}
		}
		return nil
	})
}

// BoneMeal attempts to affect the block using a bone meal item.
func (g Fern) BoneMeal(pos cube.Pos, w *world.World) bool {
	upper := DoubleTallGrass{Type: FernDoubleTallGrass(), UpperPart: true}
	if replaceableWith(w, pos.Side(cube.FaceUp), upper) {
		w.SetBlock(pos, DoubleTallGrass{Type: FernDoubleTallGrass()}, nil)
		w.SetBlock(pos.Side(cube.FaceUp), upper, nil)
		return true
	}
	return false
}

// CompostChance ...
func (g Fern) CompostChance() float64 {
	return 0.3
}

// NeighbourUpdateTick ...
func (g Fern) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !supportsVegetation(g, w.Block(pos.Side(cube.FaceDown))) {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: g})
	}
}

// HasLiquidDrops ...
func (g Fern) HasLiquidDrops() bool {
	return true
}

// UseOnBlock ...
func (g Fern) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, g)
	if !used {
		return false
	}
	if !supportsVegetation(g, w.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	place(w, pos, g, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (g Fern) EncodeItem() (name string, meta int16) {
	return "minecraft:fern", 0
}

// EncodeBlock ...
func (g Fern) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:fern", nil
}
