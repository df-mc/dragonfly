package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// BrownMushroom is a variety of fungus that grows and spreads in dark areas.
// TODO: Spreading and growth from mushrooms
type BrownMushroom struct {
	empty
	transparent
}

// LightEmissionLevel returns 1.
func (BrownMushroom) LightEmissionLevel() uint8 {
	return 1
}

// BreakInfo ...
func (b BrownMushroom) BreakInfo() BreakInfo {
	return BreakInfo{
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       simpleDrops(item.NewStack(b, 1)),
	}
}

// UseOnBlock ...
func (b BrownMushroom) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, b)
	if !used {
		return false
	}

	blockBelow := w.Block(pos.Side(cube.FaceDown))
	if diffuser, ok := blockBelow.(LightDiffuser); ok {
		if diffuser.LightDiffusionLevel() == 0 {
			return false
		}
	}

	place(w, pos, b, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (BrownMushroom) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	blockBelow := w.Block(pos.Side(cube.FaceDown))
	if diffuser, ok := blockBelow.(LightDiffuser); ok {
		if diffuser.LightDiffusionLevel() == 0 {
			w.BreakBlock(pos)
		}
	}
}

// EncodeItem ...
func (BrownMushroom) EncodeItem() (name string, meta int16) {
	return "minecraft:brown_mushroom", 0
}

// EncodeBlock ...
func (BrownMushroom) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:brown_mushroom", map[string]interface{}{}
}
