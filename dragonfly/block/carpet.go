package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/colour"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Carpet is a colourful block that can be obtained by killing/shearing sheep, or crafted using four string.
type Carpet struct {
	Colour colour.Colour
}

// CanDisplace ...
func (Carpet) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (Carpet) SideClosed(world.BlockPos, world.BlockPos, *world.World) bool {
	return false
}

// AABB ...
func (Carpet) AABB(world.BlockPos, *world.World) []physics.AABB {
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.0625, 1})}
}

// BreakInfo ...
func (c Carpet) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.1,
		Harvestable: alwaysHarvestable,
		Effective:   neverEffective,
		Drops:       simpleDrops(item.NewStack(c, 1)),
	}
}

// EncodeItem ...
func (c Carpet) EncodeItem() (id int32, meta int16) {
	return 171, int16(c.Colour.Uint8())
}

// EncodeBlock ...
func (c Carpet) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:carpet", map[string]interface{}{"color": c.Colour.String()}
}

// Hash ...
func (c Carpet) Hash() uint64 {
	return hashCarpet | (uint64(c.Colour.Uint8()) << 32)
}

// HasLiquidDrops ...
func (Carpet) HasLiquidDrops() bool {
	return true
}

// NeighbourUpdateTick ...
func (Carpet) NeighbourUpdateTick(pos, changed world.BlockPos, w *world.World) {
	if _, ok := w.Block(pos.Add(world.BlockPos{0, -1})).(Air); ok {
		w.ScheduleBlockUpdate(pos, time.Second/20)
	}
}

// ScheduledTick ...
func (Carpet) ScheduledTick(pos world.BlockPos, w *world.World) {
	if _, ok := w.Block(pos.Add(world.BlockPos{0, -1})).(Air); ok {
		w.BreakBlock(pos)
	}
}

// UseOnBlock handles not placing carpets on top of air blocks.
func (c Carpet) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, wrld *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(wrld, pos, face, c)
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
