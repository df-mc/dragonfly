package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// WoodDoor is a block that can be used as an openable 1x2 barrier.
type WoodDoor struct {
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

// FlammabilityInfo ...
func (d WoodDoor) FlammabilityInfo() FlammabilityInfo {
	if !woodTypeFlammable(d.Wood) {
		return FlammabilityInfo{}
	}
	return FlammabilityInfo{LavaFlammable: true}
}

// Model ...
func (d WoodDoor) Model() world.BlockModel {
	return model.Door{Facing: d.Facing, Open: d.Open, Right: d.Right}
}

// NeighbourUpdateTick ...
func (d WoodDoor) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if d.Top {
		if _, ok := w.Block(pos.Side(world.FaceDown)).(WoodDoor); !ok {
			w.BreakBlock(pos)
		}
	} else {
		if solid := w.Block(pos.Side(world.FaceDown)).Model().FaceSolid(pos.Side(world.FaceDown), world.FaceUp, w); !solid {
			w.BreakBlock(pos)
		} else if _, ok := w.Block(pos.Side(world.FaceUp)).(WoodDoor); !ok {
			w.BreakBlock(pos)
		}
	}
}

// UseOnBlock handles the directional placing of doors
func (d WoodDoor) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
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
	if door, ok := left.(WoodDoor); ok {
		if door.Wood == d.Wood {
			d.Right = true
		}
	}
	// The side the door hinge is on can be affected by the blocks to the left and right of the door. In particular,
	// opaque blocks on the right side of the door with transparent blocks on the left side result in a right sided
	// door hinge.
	if diffuser, ok := right.(LightDiffuser); !ok || diffuser.LightDiffusionLevel() != 0 {
		if diffuser, ok := left.(LightDiffuser); ok && diffuser.LightDiffusionLevel() == 0 {
			d.Right = true
		}
	}

	ctx.IgnoreAABB = true
	place(w, pos, d, user, ctx)
	place(w, pos.Side(world.FaceUp), WoodDoor{Wood: d.Wood, Facing: d.Facing, Top: true, Right: d.Right}, user, ctx)
	return placed(ctx)
}

// Activate ...
func (d WoodDoor) Activate(pos world.BlockPos, _ world.Face, w *world.World, _ item.User) {
	d.Open = !d.Open
	w.PlaceBlock(pos, d)

	otherPos := pos.Side(world.Face(boolByte(!d.Top)))
	other := w.Block(otherPos)
	if door, ok := other.(WoodDoor); ok {
		door.Open = d.Open
		w.PlaceBlock(otherPos, door)
	}

	w.PlaySound(pos.Vec3Centre(), sound.Door{})
}

// BreakInfo ...
func (d WoodDoor) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(d, 1)),
	}
}

// CanDisplace ...
func (d WoodDoor) CanDisplace(l world.Liquid) bool {
	_, water := l.(Water)
	return water
}

// SideClosed ...
func (d WoodDoor) SideClosed(world.BlockPos, world.BlockPos, *world.World) bool {
	return false
}

// EncodeItem ...
func (d WoodDoor) EncodeItem() (id int32, meta int16) {
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
	case wood.Crimson():
		return 755, 0
	case wood.Warped():
		return 756, 0
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (d WoodDoor) EncodeBlock() (name string, properties map[string]interface{}) {
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
	default:
		return "minecraft:" + d.Wood.String() + "_door", map[string]interface{}{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
	}
}

// Hash ...
func (d WoodDoor) Hash() uint64 {
	return hashDoor | (uint64(d.Facing) << 32) | (uint64(boolByte(d.Right)) << 35) | (uint64(boolByte(d.Open)) << 36) | (uint64(boolByte(d.Top)) << 37) | (uint64(d.Wood.Uint8()) << 38)
}

// allDoors returns a list of all door types
func allDoors() (doors []world.Block) {
	for _, w := range wood.All() {
		for i := world.Direction(0); i <= 3; i++ {
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: false, Top: false, Right: false})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: false, Top: true, Right: false})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: true, Top: true, Right: false})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: true, Top: false, Right: false})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: false, Top: false, Right: true})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: false, Top: true, Right: true})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: true, Top: true, Right: true})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: true, Top: false, Right: true})
		}
	}
	return
}
