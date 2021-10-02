package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RedMushroom is a variety of fungus that grows and spreads in dark areas.
// TODO: Spreading and growth from mushrooms
type RedMushroom struct {
	empty
	transparent
}

// BreakInfo ...
func (r RedMushroom) BreakInfo() BreakInfo {
	return BreakInfo{
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       simpleDrops(item.NewStack(r, 1)),
	}
}

// UseOnBlock ...
func (r RedMushroom) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, r)
	if !used {
		return false
	}

	blockBelow := w.Block(pos.Side(cube.FaceDown))
	if diffuser, ok := blockBelow.(LightDiffuser); ok {
		if diffuser.LightDiffusionLevel() == 0 {
			return false
		}
	}

	place(w, pos, r, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (RedMushroom) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	blockBelow := w.Block(pos.Side(cube.FaceDown))
	if diffuser, ok := blockBelow.(LightDiffuser); ok {
		if diffuser.LightDiffusionLevel() == 0 {
			w.BreakBlock(pos)
		}
	}
}

// EncodeItem ...
func (RedMushroom) EncodeItem() (name string, meta int16) {
	return "minecraft:red_mushroom", 0
}

// EncodeBlock ...
func (RedMushroom) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:red_mushroom", map[string]interface{}{}
}
