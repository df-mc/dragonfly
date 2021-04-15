package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Basalt is a type of igneous rock found in the Nether.
type Basalt struct {
	noNBT
	solid
	bassDrum

	// Polished specifies if the basalt is its polished variant.
	Polished bool
	// Axis is the axis which the basalt faces.
	Axis cube.Axis
}

// UseOnBlock ...
func (b Basalt) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, b)
	if !used {
		return
	}
	b.Axis = face.Axis()

	place(w, pos, b, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (b Basalt) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    1.25,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(b, 1)),
	}
}

// EncodeItem ...
func (b Basalt) EncodeItem() (id int32, meta int16) {
	if b.Polished {
		return -235, 0
	}
	return -234, 0
}

// EncodeBlock ...
func (b Basalt) EncodeBlock() (name string, properties map[string]interface{}) {
	if b.Polished {
		return "minecraft:polished_basalt", map[string]interface{}{"pillar_axis": b.Axis.String()}
	}
	return "minecraft:basalt", map[string]interface{}{"pillar_axis": b.Axis.String()}
}

// Hash ...
func (b Basalt) Hash() uint64 {
	return hashBasalt | (uint64(boolByte(b.Polished)) << 32) | (uint64(b.Axis) << 33)
}

// allBasalt ...
func allBasalt() (basalt []canEncode) {
	for _, axis := range cube.Axes() {
		basalt = append(basalt, Basalt{Axis: axis, Polished: false})
		basalt = append(basalt, Basalt{Axis: axis, Polished: true})
	}
	return
}
