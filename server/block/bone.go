package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Bone is a decorative block that can face different directions.
type Bone struct {
	solid

	// Axis is the axis which the bone block faces.
	Axis cube.Axis
}

func (b Bone) Instrument() sound.Instrument {
	return sound.Xylophone()
}

// UseOnBlock handles the rotational placing of bone blocks.
func (b Bone) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, b)
	if !used {
		return
	}
	b.Axis = face.Axis()

	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

func (b Bone) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(b))
}

func (b Bone) EncodeItem() (name string, meta int16) {
	return "minecraft:bone_block", 0
}

func (b Bone) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:bone_block", map[string]any{"pillar_axis": b.Axis.String(), "deprecated": int32(0)}
}

func allBoneBlock() (boneBlocks []world.Block) {
	for _, axis := range cube.Axes() {
		boneBlocks = append(boneBlocks, Bone{Axis: axis})
	}
	return
}
