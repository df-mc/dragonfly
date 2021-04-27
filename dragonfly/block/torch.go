package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/block/fire"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Torch are non-solid blocks that emit light.
type Torch struct {
	noNBT
	transparent
	empty

	// Facing is the direction from the torch to the block.
	Facing cube.Face
	// Type is the type of fire lighting the torch.
	Type fire.Fire
}

// LightEmissionLevel ...
func (t Torch) LightEmissionLevel() uint8 {
	switch t.Type {
	case fire.Normal():
		return 14
	default:
		return t.Type.LightLevel
	}
}

// UseOnBlock ...
func (t Torch) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, t)
	if !used {
		return false
	}
	if face == cube.FaceDown {
		return false
	}
	if _, ok := w.Block(pos).(world.Liquid); ok {
		return false
	}
	if !w.Block(pos.Side(face.Opposite())).Model().FaceSolid(pos.Side(face.Opposite()), face, w) {
		found := false
		for _, i := range []cube.Face{cube.FaceSouth, cube.FaceWest, cube.FaceNorth, cube.FaceEast, cube.FaceDown} {
			if w.Block(pos.Side(i)).Model().FaceSolid(pos.Side(i), i.Opposite(), w) {
				found = true
				face = i.Opposite()
				break
			}
		}
		if !found {
			return false
		}
	}
	t.Facing = face.Opposite()

	place(w, pos, t, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (t Torch) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !w.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), w) {
		w.BreakBlockWithoutParticles(pos)
	}
}

// HasLiquidDrops ...
func (t Torch) HasLiquidDrops() bool {
	return true
}

// EncodeItem ...
func (t Torch) EncodeItem() (id int32, meta int16) {
	switch t.Type {
	case fire.Normal():
		return 50, 0
	case fire.Soul():
		return -268, 0
	}
	panic("invalid fire type")
}

// EncodeBlock ...
func (t Torch) EncodeBlock() (name string, properties map[string]interface{}) {
	facing := "unknown"
	switch t.Facing {
	case cube.FaceDown:
		facing = "top"
	case cube.FaceNorth:
		facing = "north"
	case cube.FaceEast:
		facing = "east"
	case cube.FaceSouth:
		facing = "south"
	case cube.FaceWest:
		facing = "west"
	}

	switch t.Type {
	case fire.Normal():
		return "minecraft:torch", map[string]interface{}{"torch_facing_direction": facing}
	case fire.Soul():
		return "minecraft:soul_torch", map[string]interface{}{"torch_facing_direction": facing}
	}
	panic("invalid fire type")
}

// allTorch ...
func allTorch() (torch []world.Block) {
	for i := cube.Face(0); i < 6; i++ {
		torch = append(torch, Torch{Type: fire.Normal(), Facing: i})
		torch = append(torch, Torch{Type: fire.Soul(), Facing: i})
	}
	return
}
