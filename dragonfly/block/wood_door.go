package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// WoodDoor is a block that can be used as an openable 1x2 barrier.
type WoodDoor struct {
	transparent
	bass

	// Wood is the type of wood of the door. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// Facing is the direction the door is facing.
	Facing cube.Direction
	// Open is whether or not the door is open.
	Open bool
	// Top is whether the block is the top or bottom half of a door
	Top bool
	// Right is whether the door hinge is on the right side
	Right bool
}

// FlammabilityInfo ...
func (d WoodDoor) FlammabilityInfo() FlammabilityInfo {
	if !d.Wood.Flammable() {
		return FlammabilityInfo{}
	}
	return FlammabilityInfo{LavaFlammable: true}
}

// Model ...
func (d WoodDoor) Model() world.BlockModel {
	return model.Door{Facing: d.Facing, Open: d.Open, Right: d.Right}
}

// NeighbourUpdateTick ...
func (d WoodDoor) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if d.Top {
		if _, ok := w.Block(pos.Side(cube.FaceDown)).(WoodDoor); !ok {
			w.BreakBlock(pos)
		}
	} else {
		if solid := w.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceUp, w); !solid {
			w.BreakBlock(pos)
		} else if _, ok := w.Block(pos.Side(cube.FaceUp)).(WoodDoor); !ok {
			w.BreakBlock(pos)
		}
	}
}

// UseOnBlock handles the directional placing of doors
func (d WoodDoor) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, d)
	if !used {
		return false
	}
	if face != cube.FaceUp {
		return false
	}
	if solid := w.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceUp, w); !solid {
		return false
	}
	if _, ok := w.Block(pos.Side(cube.FaceUp)).(Air); !ok {
		return false
	}
	d.Facing = user.Facing()
	left := w.Block(pos.Side(d.Facing.RotateLeft90().Face()))
	right := w.Block(pos.Side(d.Facing.RotateRight90().Face()))
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
	place(w, pos.Side(cube.FaceUp), WoodDoor{Wood: d.Wood, Facing: d.Facing, Top: true, Right: d.Right}, user, ctx)
	return placed(ctx)
}

// Activate ...
func (d WoodDoor) Activate(pos cube.Pos, _ cube.Face, w *world.World, _ item.User) {
	d.Open = !d.Open
	w.PlaceBlock(pos, d)

	otherPos := pos.Side(cube.Face(boolByte(!d.Top)))
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
func (d WoodDoor) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// EncodeItem ...
func (d WoodDoor) EncodeItem() (id int32, name string, meta int16) {
	switch d.Wood {
	case wood.Oak():
		return 324, "minecraft:wooden_door", 0
	case wood.Spruce():
		return 427, "minecraft:spruce_door", 0
	case wood.Birch():
		return 428, "minecraft:birch_door", 0
	case wood.Jungle():
		return 429, "minecraft:jungle_door", 0
	case wood.Acacia():
		return 430, "minecraft:acacia_door", 0
	case wood.DarkOak():
		return 431, "minecraft:dark_oak_door", 0
	case wood.Crimson():
		return 755, "minecraft:crimson_door", 0
	case wood.Warped():
		return 756, "minecraft:warped_door", 0
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (d WoodDoor) EncodeBlock() (name string, properties map[string]interface{}) {
	direction := 3
	switch d.Facing {
	case cube.South:
		direction = 1
	case cube.West:
		direction = 2
	case cube.East:
		direction = 0
	}

	switch d.Wood {
	case wood.Oak():
		return "minecraft:wooden_door", map[string]interface{}{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
	default:
		return "minecraft:" + d.Wood.String() + "_door", map[string]interface{}{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
	}
}

// allDoors returns a list of all door types
func allDoors() (doors []world.Block) {
	for _, w := range wood.All() {
		for i := cube.Direction(0); i <= 3; i++ {
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
