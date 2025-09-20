package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Stonecutter is used to craft stone and copper related blocks in smaller and more precise quantities than crafting,
// and is more efficient than crafting for certain recipes.
type Stonecutter struct {
	bassDrum

	// Facing is the direction the stonecutter is facing.
	Facing cube.Direction
}

func (Stonecutter) Model() world.BlockModel {
	return model.Stonecutter{}
}

func (s Stonecutter) BreakInfo() BreakInfo {
	return newBreakInfo(3.5, pickaxeHarvestable, pickaxeEffective, oneOf(s))
}

func (Stonecutter) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos, tx)
		return true
	}
	return false
}

func (s Stonecutter) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, s)
	if !used {
		return
	}
	s.Facing = user.Rotation().Direction().Opposite()
	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

func (Stonecutter) EncodeItem() (name string, meta int16) {
	return "minecraft:stonecutter_block", 0
}

func (s Stonecutter) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:stonecutter_block", map[string]any{"minecraft:cardinal_direction": s.Facing.String()}
}

func allStonecutters() (stonecutters []world.Block) {
	for _, d := range cube.Directions() {
		stonecutters = append(stonecutters, Stonecutter{Facing: d})
	}
	return
}
