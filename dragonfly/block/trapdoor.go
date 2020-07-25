package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// Trapdoor is a block that can be used as an openable 1x1 barrier
type Trapdoor struct {
	noNBT

	// Wood is the type of wood of the trapdoor. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// Facing is the direction the trapdoor is facing.
	Facing world.Direction
	// Open is whether or not the trapdoor is open.
	Open bool
	// Top is whether the trapdoor occupies the top or bottom part of a block.
	Top bool
}

// LightDiffusionLevel ...
func (t Trapdoor) LightDiffusionLevel() uint8 {
	return 0
}

// AABB ...
func (t Trapdoor) AABB(pos world.BlockPos, w *world.World) []physics.AABB {
	if t.Open {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(int(t.Facing.Face()), -0.8125)}
	}
	if t.Top {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(int(world.FaceDown), -0.8125)}
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(int(world.FaceUp), -0.8125)}
}

// UseOnBlock handles the directional placing of trapdoors and makes sure they are properly placed upside down
// when needed.
func (t Trapdoor) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, t)
	if !used {
		return false
	}
	t.Facing = user.Facing().Opposite()
	t.Top = (clickPos.Y() > 0.5 && face != world.FaceUp) || face == world.FaceDown

	place(w, pos, t, user, ctx)
	return placed(ctx)
}

// Activate ...
func (t Trapdoor) Activate(pos world.BlockPos, clickedFace world.Face, w *world.World, u item.User) {
	t.Open = !t.Open
	w.SetBlock(pos, t)
	w.PlaySound(pos.Vec3Centre(), sound.Door{})
}

// BreakInfo ...
func (t Trapdoor) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(t, 1)),
	}
}

// CanDisplace ...
func (t Trapdoor) CanDisplace(l world.Liquid) bool {
	_, water := l.(Water)
	return water
}

// SideClosed ...
func (t Trapdoor) SideClosed(pos, side world.BlockPos, w *world.World) bool {
	return false
}

// EncodeItem ...
func (t Trapdoor) EncodeItem() (id int32, meta int16) {
	switch t.Wood {
	case wood.Oak():
		return 96, 0
	case wood.Spruce():
		return -149, 0
	case wood.Birch():
		return -146, 0
	case wood.Jungle():
		return -148, 0
	case wood.Acacia():
		return -145, 0
	case wood.DarkOak():
		return -147, 0
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (t Trapdoor) EncodeBlock() (name string, properties map[string]interface{}) {
	switch t.Wood {
	case wood.Oak():
		return "minecraft:trapdoor", map[string]interface{}{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
	case wood.Spruce():
		return "minecraft:spruce_trapdoor", map[string]interface{}{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
	case wood.Birch():
		return "minecraft:birch_trapdoor", map[string]interface{}{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
	case wood.Jungle():
		return "minecraft:jungle_trapdoor", map[string]interface{}{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
	case wood.Acacia():
		return "minecraft:acacia_trapdoor", map[string]interface{}{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
	case wood.DarkOak():
		return "minecraft:dark_oak_trapdoor", map[string]interface{}{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
	}
	panic("invalid wood type")
}

// Hash ...
func (t Trapdoor) Hash() uint64 {
	return hashTrapdoor | (uint64(t.Facing) << 32) | (uint64(boolByte(t.Open)) << 34) | (uint64(boolByte(t.Top)) << 35) | (uint64(t.Wood.Uint8()) << 36)
}

// allTrapdoors returns a list of all trapdoor types
func allTrapdoors() (trapdoors []world.Block) {
	for _, w := range []wood.Wood{
		wood.Oak(),
		wood.Spruce(),
		wood.Birch(),
		wood.Jungle(),
		wood.Acacia(),
		wood.DarkOak(),
	} {
		for i := world.Direction(0); i <= 3; i++ {
			trapdoors = append(trapdoors, Trapdoor{Wood: w, Facing: i, Open: false, Top: false})
			trapdoors = append(trapdoors, Trapdoor{Wood: w, Facing: i, Open: false, Top: true})
			trapdoors = append(trapdoors, Trapdoor{Wood: w, Facing: i, Open: true, Top: true})
			trapdoors = append(trapdoors, Trapdoor{Wood: w, Facing: i, Open: true, Top: false})
		}
	}
	return
}
