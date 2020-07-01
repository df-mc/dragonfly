package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/colour"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// AABB ...
func (Carpet) AABB(world.BlockPos, *world.World) []physics.AABB {
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.0625, 1})}
}

// BreakInfo ...
func (w Carpet) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.1,
		Harvestable: alwaysHarvestable,
		Effective:   alwaysEffective,
		Drops:       simpleDrops(item.NewStack(w, 1)),
	}
}

// Carpet is a colourful block that can be obtained by killing/shearing sheep, or crafted using four string.
type Carpet struct {
	Colour colour.Colour
}

// EncodeItem ...
func (w Carpet) EncodeItem() (id int32, meta int16) {
	return 171, int16(w.Colour.Uint8())
}

// EncodeBlock ...
func (w Carpet) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:carpet", map[string]interface{}{"color": w.Colour.String()}
}

// UseOnBlock handles not placing carpets on top of air blocks.
func (c Carpet) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, wrld *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(wrld, pos, face, c)
	if !used {
		return
	}

	if _, ok := wrld.Block((world.BlockPos{pos.X(), pos.Y() - 1, pos.Z()})).(Air); ok {
		return
	}

	place(wrld, pos, c, user, ctx)
	return placed(ctx)
}

// allCarpets returns carpet blocks with all possible colours.
func allCarpets() []world.Block {
	b := make([]world.Block, 0, 16)
	for _, c := range colour.All() {
		b = append(b, Carpet{Colour: c})
	}
	return b
}
