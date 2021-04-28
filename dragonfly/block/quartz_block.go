package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

type (
	// Quartz is a mineral block used only for decoration.
	Quartz struct {
		solid
		bassDrum
		// Smooth specifies if the quartz block is smooth or not.
		Smooth bool
	}

	// ChiseledQuartz is a mineral block used only for decoration.
	ChiseledQuartz struct {
		solid
		bassDrum
	}
	// QuartzPillar is a mineral block used only for decoration.
	QuartzPillar struct {
		solid
		bassDrum
		// Axis is the axis which the quartz pillar block faces.
		Axis cube.Axis
	}
)

// UseOnBlock handles the rotational placing of quartz pillar blocks.
func (q QuartzPillar) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, q)
	if !used {
		return
	}
	q.Axis = face.Axis()

	place(w, pos, q, user, ctx)
	return placed(ctx)
}

var quartzBreakInfo = BreakInfo{
	Hardness:    0.8,
	Harvestable: pickaxeHarvestable,
	Effective:   pickaxeEffective,
	Drops:       simpleDrops(item.NewStack(Quartz{}, 1)),
}

// BreakInfo ...
func (q Quartz) BreakInfo() BreakInfo {
	return quartzBreakInfo
}

// BreakInfo ...
func (c ChiseledQuartz) BreakInfo() BreakInfo {
	i := quartzBreakInfo
	i.Drops = simpleDrops(item.NewStack(c, 1))
	return i
}

// BreakInfo ...
func (q QuartzPillar) BreakInfo() BreakInfo {
	i := quartzBreakInfo
	i.Drops = simpleDrops(item.NewStack(q, 1))
	return i
}

// EncodeItem ...
func (q Quartz) EncodeItem() (id int32, meta int16) {
	if q.Smooth {
		return 155, 3
	}
	return 155, 0
}

// EncodeItem ...
func (c ChiseledQuartz) EncodeItem() (id int32, meta int16) {
	return 155, 1
}

// EncodeItem ...
func (q QuartzPillar) EncodeItem() (id int32, meta int16) {
	return 155, 2
}

// EncodeBlock ...
func (q Quartz) EncodeBlock() (name string, properties map[string]interface{}) {
	if q.Smooth {
		return "minecraft:quartz_block", map[string]interface{}{"chisel_type": "smooth", "pillar_axis": "x"}
	}
	return "minecraft:quartz_block", map[string]interface{}{"chisel_type": "default", "pillar_axis": "x"}
}

// EncodeBlock ...
func (ChiseledQuartz) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:quartz_block", map[string]interface{}{"chisel_type": "chiseled", "pillar_axis": "x"}
}

// EncodeBlock ...
func (q QuartzPillar) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:quartz_block", map[string]interface{}{"pillar_axis": q.Axis.String(), "chisel_type": "lines"}
}

// allQuartz ...
func allQuartz() (quartz []world.Block) {
	quartz = append(quartz, Quartz{})
	quartz = append(quartz, Quartz{Smooth: true})
	quartz = append(quartz, ChiseledQuartz{})
	for _, a := range cube.Axes() {
		quartz = append(quartz, QuartzPillar{Axis: a})
	}
	return
}
