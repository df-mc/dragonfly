package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// ShortGrass is a transparent plant block which can be used to obtain seeds and as decoration.
type ShortGrass struct {
	replaceable
	transparent
	empty

	Double bool
}

// FlammabilityInfo ...
func (g ShortGrass) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// BreakInfo ...
func (g ShortGrass) BreakInfo() BreakInfo {
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
func (g ShortGrass) BoneMeal(pos cube.Pos, w *world.World) bool {
	upper := DoubleTallGrass{Type: NormalDoubleTallGrass(), UpperPart: true}
	if replaceableWith(w, pos.Side(cube.FaceUp), upper) {
		w.SetBlock(pos, DoubleTallGrass{Type: NormalDoubleTallGrass()}, nil)
		w.SetBlock(pos.Side(cube.FaceUp), upper, nil)
		return true
	}
	return false
}

// CompostChance ...
func (g ShortGrass) CompostChance() float64 {
	return 0.65
}

// NeighbourUpdateTick ...
func (g ShortGrass) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !supportsVegetation(g, w.Block(pos.Side(cube.FaceDown))) {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: g})
	}
}

// HasLiquidDrops ...
func (g ShortGrass) HasLiquidDrops() bool {
	return true
}

// UseOnBlock ...
func (g ShortGrass) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
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
func (g ShortGrass) EncodeItem() (name string, meta int16) {
	return "minecraft:short_grass", 0
}

// EncodeBlock ...
func (g ShortGrass) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:short_grass", nil
}
