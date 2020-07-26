package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Door is a block that can be used as an openable 1x2 barrier.
type Door struct {
	noNBT
	transparent

	// Wood is the type of wood of the door. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// Facing is the direction the door is facing.
	Facing world.Direction
	// Open is whether or not the door is open.
	Open bool
	// Top is whether the block is the top or bottom half of a door
	Top bool
	// Right is whether the door hinge is on the right side
	Right bool
}

// Model ...
func (d Door) Model() world.BlockModel {
	return model.Door{Facing: d.Facing, Open: d.Open, Right: d.Right}
}

// NeighbourUpdateTick ...
func (d Door) NeighbourUpdateTick(pos, changedNeighbour world.BlockPos, w *world.World) {
	if d.Top {
		if _, ok := w.Block(pos.Side(world.FaceDown)).(Door); !ok {
			w.SetBlock(pos, nil)
		}
	} else {
		if solid := w.Block(pos.Side(world.FaceDown)).Model().FaceSolid(pos.Side(world.FaceDown), world.FaceUp, w); !solid {
			w.SetBlock(pos, nil)
		} else if _, ok := w.Block(pos.Side(world.FaceUp)).(Door); !ok {
			w.SetBlock(pos, nil)
		}
	}
}

// UseOnBlock handles the directional placing of doors
func (d Door) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, d)
	if !used {
		return false
	}
	if face != world.FaceUp {
		return false
	}
	if solid := w.Block(pos.Side(world.FaceDown)).Model().FaceSolid(pos.Side(world.FaceDown), world.FaceUp, w); !solid {
		return false
	}
	if _, ok := w.Block(pos.Side(world.FaceUp)).(Air); !ok {
		return false
	}
	d.Facing = user.Facing()
	left := w.Block(pos.Side(d.Facing.Rotate90().Opposite().Face()))
	right := w.Block(pos.Side(d.Facing.Rotate90().Face()))
	if door, ok := left.(Door); ok {
		if door.Wood == d.Wood {
			d.Right = true
		}
	}
	if diffuser, ok := right.(LightDiffuser); !ok || diffuser.LightDiffusionLevel() != 0 {
		if diffuser, ok := left.(LightDiffuser); ok && diffuser.LightDiffusionLevel() == 0 {
			d.Right = true
		}
	}

	place(w, pos, d, user, ctx)
	place(w, pos.Side(world.FaceUp), Door{Wood: d.Wood, Facing: d.Facing, Top: true, Right: d.Right}, user, ctx)
	return placed(ctx)
}

// Activate ...
func (d Door) Activate(pos world.BlockPos, clickedFace world.Face, w *world.World, u item.User) {
	d.Open = !d.Open
	w.SetBlock(pos, d)

	otherPos := pos.Side(world.Face(boolByte(!d.Top)))
	other := w.Block(otherPos)
	if door, ok := other.(Door); ok {
		door.Open = d.Open
		w.SetBlock(otherPos, door)
	}

	w.PlaySound(pos.Vec3Centre(), sound.Door{})
}

// BreakInfo ...
func (d Door) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(d, 1)),
	}
}

// CanDisplace ...
func (d Door) CanDisplace(l world.Liquid) bool {
	_, water := l.(Water)
	return water
}

// SideClosed ...
func (d Door) SideClosed(pos, side world.BlockPos, w *world.World) bool {
	return false
}

// EncodeItem ...
func (d Door) EncodeItem() (id int32, meta int16) {
	switch d.Wood {
	case wood.Oak():
		return 324, 0
	case wood.Spruce():
		return 427, 0
	case wood.Birch():
		return 428, 0
	case wood.Jungle():
		return 429, 0
	case wood.Acacia():
		return 430, 0
	case wood.DarkOak():
		return 431, 0
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (d Door) EncodeBlock() (name string, properties map[string]interface{}) {
	direction := 3
	switch d.Facing {
	case world.South:
		direction = 1
	case world.West:
		direction = 2
	case world.East:
		direction = 0
	}

	switch d.Wood {
	case wood.Oak():
		return "minecraft:wooden_door", map[string]interface{}{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
	case wood.Spruce():
		return "minecraft:spruce_door", map[string]interface{}{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
	case wood.Birch():
		return "minecraft:birch_door", map[string]interface{}{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
	case wood.Jungle():
		return "minecraft:jungle_door", map[string]interface{}{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
	case wood.Acacia():
		return "minecraft:acacia_door", map[string]interface{}{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
	case wood.DarkOak():
		return "minecraft:dark_oak_door", map[string]interface{}{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
	}
	panic("invalid wood type")
}

// Hash ...
func (d Door) Hash() uint64 {
	return hashDoor | (uint64(d.Facing) << 32) | (uint64(boolByte(d.Right)) << 35) | (uint64(boolByte(d.Open)) << 36) | (uint64(boolByte(d.Top)) << 37) | (uint64(d.Wood.Uint8()) << 38)
}

// allDoors returns a list of all door types
func allDoors() (doors []world.Block) {
	for _, w := range []wood.Wood{
		wood.Oak(),
		wood.Spruce(),
		wood.Birch(),
		wood.Jungle(),
		wood.Acacia(),
		wood.DarkOak(),
	} {
		for i := world.Direction(0); i <= 3; i++ {
			doors = append(doors, Door{Wood: w, Facing: i, Open: false, Top: false, Right: false})
			doors = append(doors, Door{Wood: w, Facing: i, Open: false, Top: true, Right: false})
			doors = append(doors, Door{Wood: w, Facing: i, Open: true, Top: true, Right: false})
			doors = append(doors, Door{Wood: w, Facing: i, Open: true, Top: false, Right: false})
			doors = append(doors, Door{Wood: w, Facing: i, Open: false, Top: false, Right: true})
			doors = append(doors, Door{Wood: w, Facing: i, Open: false, Top: true, Right: true})
			doors = append(doors, Door{Wood: w, Facing: i, Open: true, Top: true, Right: true})
			doors = append(doors, Door{Wood: w, Facing: i, Open: true, Top: false, Right: true})
		}
	}
	return
}
