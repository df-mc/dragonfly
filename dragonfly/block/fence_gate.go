package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// FenceGate is a block that can be used as an openable 1x1 barrier.
type FenceGate struct {
	noNBT
	transparent

	// Wood is the type of wood of the fence gate. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// Facing is the direction the fence gate swings open.
	Facing world.Direction
	// Open is whether the fence gate is open.
	Open bool
	// InWall lowers the fence gate by 3 pixels if placed next to wall blocks.
	InWall bool
}

// UseOnBlock ...
func (f FenceGate) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, f)
	if !used {
		return false
	}
	f.Facing = user.Facing()
	//TODO: Set InWall if placed next to wall block

	place(w, pos, f, user, ctx)
	return placed(ctx)
}

// Activate ...
func (f FenceGate) Activate(pos world.BlockPos, clickedFace world.Face, w *world.World, u item.User) {
	f.Open = !f.Open
	if f.Open && f.Facing.Opposite() == u.Facing() {
		f.Facing = u.Facing()
	}
	w.PlaceBlock(pos, f)
	w.PlaySound(pos.Vec3Centre(), sound.Door{})
}

// CanDisplace ...
func (f FenceGate) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (f FenceGate) SideClosed(pos, side world.BlockPos, w *world.World) bool {
	return false
}

// EncodeItem ...
func (f FenceGate) EncodeItem() (id int32, meta int16) {
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
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (f FenceGate) EncodeBlock() (name string, properties map[string]interface{}) {
	directions := map[world.Direction]int32{
		world.North: 2,
		world.South: 0,
		world.West:  1,
		world.East:  3,
	}

	switch f.Wood {
	case wood.Oak():
		return "minecraft:fence_gate", map[string]interface{}{"direction": directions[f.Facing], "open_bit": f.Open, "in_wall_bit": f.InWall}
	case wood.Spruce():
		return "minecraft:spruce_fence_gate", map[string]interface{}{"direction": directions[f.Facing], "open_bit": f.Open, "in_wall_bit": f.InWall}
	case wood.Birch():
		return "minecraft:birch_fence_gate", map[string]interface{}{"direction": directions[f.Facing], "open_bit": f.Open, "in_wall_bit": f.InWall}
	case wood.Jungle():
		return "minecraft:jungle_fence_gate", map[string]interface{}{"direction": directions[f.Facing], "open_bit": f.Open, "in_wall_bit": f.InWall}
	case wood.Acacia():
		return "minecraft:acacia_fence_gate", map[string]interface{}{"direction": directions[f.Facing], "open_bit": f.Open, "in_wall_bit": f.InWall}
	case wood.DarkOak():
		return "minecraft:dark_oak_fence_gate", map[string]interface{}{"direction": directions[f.Facing], "open_bit": f.Open, "in_wall_bit": f.InWall}
	}
	panic("invalid wood type")
}

// Hash ...
func (f FenceGate) Hash() uint64 {
	return hashFenceGate | (uint64(f.Facing) << 32) | (uint64(boolByte(f.Open)) << 34) | (uint64(boolByte(f.InWall)) << 35) | (uint64(f.Wood.Uint8()) << 36)
}

// Model ...
func (f FenceGate) Model() world.BlockModel {
	return model.FenceGate{Facing: f.Facing, Open: f.Open, InWall: f.InWall}
}

// allFenceGates returns a list of all trapdoor types.
func allFenceGates() (trapdoors []world.Block) {
	for _, w := range []wood.Wood{
		wood.Oak(),
		wood.Spruce(),
		wood.Birch(),
		wood.Jungle(),
		wood.Acacia(),
		wood.DarkOak(),
	} {
		for i := world.Direction(0); i <= 3; i++ {
			trapdoors = append(trapdoors, FenceGate{Wood: w, Facing: i, Open: false, InWall: false})
			trapdoors = append(trapdoors, FenceGate{Wood: w, Facing: i, Open: false, InWall: true})
			trapdoors = append(trapdoors, FenceGate{Wood: w, Facing: i, Open: true, InWall: true})
			trapdoors = append(trapdoors, FenceGate{Wood: w, Facing: i, Open: true, InWall: false})
		}
	}
	return
}
