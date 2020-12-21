package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// WoodFenceGate is a block that can be used as an openable 1x1 barrier.
type WoodFenceGate struct {
	noNBT
	transparent
	bass

	// Wood is the type of wood of the fence gate. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// Facing is the direction the fence gate swings open.
	Facing world.Direction
	// Open is whether the fence gate is open.
	Open bool
	// Lowered lowers the fence gate by 3 pixels and is set when placed next to wall blocks.
	Lowered bool
}

// BreakInfo ...
func (f WoodFenceGate) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(f, 1)),
	}
}

// FlammabilityInfo ...
func (f WoodFenceGate) FlammabilityInfo() FlammabilityInfo {
	if !f.Wood.Flammable() {
		return FlammabilityInfo{}
	}
	return FlammabilityInfo{
		Encouragement: 5,
		Flammability:  20,
		LavaFlammable: true,
	}
}

// UseOnBlock ...
func (f WoodFenceGate) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, f)
	if !used {
		return false
	}
	f.Facing = user.Facing()
	//TODO: Set Lowered if placed next to wall block

	place(w, pos, f, user, ctx)
	return placed(ctx)
}

// Activate ...
func (f WoodFenceGate) Activate(pos world.BlockPos, clickedFace world.Face, w *world.World, u item.User) {
	f.Open = !f.Open
	if f.Open && f.Facing.Opposite() == u.Facing() {
		f.Facing = u.Facing()
	}
	w.PlaceBlock(pos, f)
	w.PlaySound(pos.Vec3Centre(), sound.Door{})
}

// CanDisplace ...
func (f WoodFenceGate) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (f WoodFenceGate) SideClosed(pos, side world.BlockPos, w *world.World) bool {
	return false
}

// EncodeItem ...
func (f WoodFenceGate) EncodeItem() (id int32, meta int16) {
	switch f.Wood {
	case wood.Oak():
		return 107, 0
	case wood.Spruce():
		return 183, 0
	case wood.Birch():
		return 184, 0
	case wood.Jungle():
		return 185, 0
	case wood.Acacia():
		return 187, 0
	case wood.DarkOak():
		return 186, 0
	case wood.Crimson():
		return -258, 0
	case wood.Warped():
		return -259, 0
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (f WoodFenceGate) EncodeBlock() (name string, properties map[string]interface{}) {
	direction := 2
	switch f.Facing {
	case world.South:
		direction = 0
	case world.West:
		direction = 1
	case world.East:
		direction = 3
	}

	switch f.Wood {
	case wood.Oak():
		return "minecraft:fence_gate", map[string]interface{}{"direction": int32(direction), "open_bit": f.Open, "in_wall_bit": f.Lowered}
	default:
		return "minecraft:" + f.Wood.String() + "_fence_gate", map[string]interface{}{"direction": int32(direction), "open_bit": f.Open, "in_wall_bit": f.Lowered}
	}
}

// Hash ...
func (f WoodFenceGate) Hash() uint64 {
	return hashFenceGate | (uint64(f.Facing) << 32) | (uint64(boolByte(f.Open)) << 34) | (uint64(boolByte(f.Lowered)) << 35) | (uint64(f.Wood.Uint8()) << 36)
}

// Model ...
func (f WoodFenceGate) Model() world.BlockModel {
	return model.FenceGate{Facing: f.Facing, Open: f.Open}
}
