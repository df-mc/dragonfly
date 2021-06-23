package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// DoubleTallGrass is a two-block high variety of grass.
type DoubleTallGrass struct {
	transparent
	replaceable
	empty

	// UpperPart is set if the plant is the upper part.
	UpperPart bool
	// Type is the type of grass
	Type GrassType
}

// HasLiquidDrops ...
func (d DoubleTallGrass) HasLiquidDrops() bool {
	return true
}

// NeighbourUpdateTick ...
func (d DoubleTallGrass) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if d.UpperPart {
		if bottom, ok := w.Block(pos.Side(cube.FaceDown)).(DoubleTallGrass); !ok || bottom.Type != d.Type || bottom.UpperPart {
			w.BreakBlock(pos)
		}
		return
	}
	if upper, ok := w.Block(pos.Side(cube.FaceUp)).(DoubleTallGrass); !ok || upper.Type != d.Type || !upper.UpperPart {
		w.BreakBlock(pos)
		return
	}
	if !supportsVegetation(w.Block(pos.Side(cube.FaceDown))) {
		w.BreakBlock(pos)
	}
}

// UseOnBlock ...
func (d DoubleTallGrass) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, d)
	if !used {
		return false
	}
	if !replaceableWith(w, pos.Side(cube.FaceUp), d) {
		return false
	}
	if !supportsVegetation(w.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	place(w, pos, d, user, ctx)
	place(w, pos.Side(cube.FaceUp), DoubleTallGrass{Type: d.Type, UpperPart: true}, user, ctx)
	return placed(ctx)
}

// FlammabilityInfo ...
func (d DoubleTallGrass) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, true)
}

// BreakInfo ...
func (d DoubleTallGrass) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, func(t tool.Tool) []item.Stack {
		if t.ToolType() == tool.TypeShears { //TODO: Silk Touch
			return []item.Stack{item.NewStack(d, 1)}
		}
		if rand.Float32() > 0.57 {
			return []item.Stack{item.NewStack(WheatSeeds{}, 1)}
		}
		return []item.Stack{}
	})
}

// EncodeItem ...
func (d DoubleTallGrass) EncodeItem() (name string, meta int16) {
	return "minecraft:double_plant", int16(d.Type.Uint8() + 2)
}

// EncodeBlock ...
func (d DoubleTallGrass) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:double_plant", map[string]interface{}{"double_flower_type": d.Type.String(), "upper_block_bit": d.UpperPart}
}

// allDoubleTallGrass ...
func allDoubleTallGrass() (b []world.Block) {
	for _, g := range GrassTypes() {
		b = append(b, DoubleTallGrass{Type: g})
		b = append(b, DoubleTallGrass{Type: g, UpperPart: true})
	}
	return
}
