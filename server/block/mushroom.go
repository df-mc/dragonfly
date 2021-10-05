package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Mushroom is a variety of fungus that grows and spreads in dark areas.
// TODO: Spreading and growth from bone meal
type Mushroom struct {
	// Type is the mushroom type. This is either brown or red.
	Type MushroomType

	empty
	transparent
}

// LightEmissionLevel returns 1.
func (m Mushroom) LightEmissionLevel() uint8 {
	if m.Type == RedMushroom() {
		return 0
	}
	return 1
}

// BreakInfo ...
func (m Mushroom) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, simpleDrops(item.NewStack(m, 1)))
}

// UseOnBlock ...
func (m Mushroom) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, m)
	if !used {
		return false
	}

	blockBelow := w.Block(pos.Side(cube.FaceDown))
	if diffuser, ok := blockBelow.(LightDiffuser); ok {
		if diffuser.LightDiffusionLevel() == 0 {
			return false
		}
	}

	place(w, pos, m, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (Mushroom) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	blockBelow := w.Block(pos.Side(cube.FaceDown))
	if diffuser, ok := blockBelow.(LightDiffuser); ok {
		if diffuser.LightDiffusionLevel() == 0 {
			w.BreakBlock(pos)
		}
	}
}

// EncodeItem ...
func (m Mushroom) EncodeItem() (name string, meta int16) {
	return "minecraft:" + m.Type.String() + "_mushroom", 0
}

// EncodeBlock ...
func (m Mushroom) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:" + m.Type.String() + "_mushroom", nil
}
