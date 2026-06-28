package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// BambooBlock is a rotatable flammable block made from bamboo.
type BambooBlock struct {
	solid
	bass

	// Axis is the axis which the bamboo block faces.
	Axis cube.Axis
	// Stripped specifies if the bamboo block is stripped.
	Stripped bool
}

// FlammabilityInfo ...
func (BambooBlock) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(5, 5, true)
}

// BreakInfo ...
func (b BambooBlock) BreakInfo() BreakInfo {
	return newBreakInfo(2.0, alwaysHarvestable, axeEffective, oneOf(b))
}

// FuelInfo ...
func (BambooBlock) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// UseOnBlock ...
func (b BambooBlock) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, b)
	if !used {
		return
	}
	b.Axis = face.Axis()

	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// Strip ...
func (b BambooBlock) Strip() (world.Block, world.Sound, bool) {
	return BambooBlock{Axis: b.Axis, Stripped: true}, nil, !b.Stripped
}

// EncodeItem ...
func (b BambooBlock) EncodeItem() (name string, meta int16) {
	if b.Stripped {
		return "minecraft:stripped_bamboo_block", 0
	}
	return "minecraft:bamboo_block", 0
}

// EncodeBlock ...
func (b BambooBlock) EncodeBlock() (name string, properties map[string]any) {
	if b.Stripped {
		return "minecraft:stripped_bamboo_block", map[string]any{"pillar_axis": b.Axis.String()}
	}
	return "minecraft:bamboo_block", map[string]any{"pillar_axis": b.Axis.String()}
}

// allBambooBlocks ...
func allBambooBlocks() (blocks []world.Block) {
	for _, axis := range cube.Axes() {
		blocks = append(blocks, BambooBlock{Axis: axis})
		blocks = append(blocks, BambooBlock{Axis: axis, Stripped: true})
	}
	return
}
