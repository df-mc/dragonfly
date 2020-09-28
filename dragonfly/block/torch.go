package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/fire"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Torch are non-solid blocks that emit light.
type Torch struct {
	noNBT
	empty

	// Facing is the direction from the torch to the block.
	Facing world.Face
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
func (t Torch) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, t)
	if !used {
		return false
	}
	if face == world.FaceDown {
		return false
	}
	if _, ok := w.Block(pos).(world.Liquid); ok {
		return false
	}
	if !w.Block(pos.Side(face.Opposite())).Model().FaceSolid(pos.Side(face.Opposite()), face, w) {
		found := false
		for _, i := range []world.Face{world.FaceSouth, world.FaceWest, world.FaceNorth, world.FaceEast, world.FaceDown} {
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
func (t Torch) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
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
	facing := "up"
	switch t.Facing {
	case world.FaceDown:
		facing = "top"
	case world.FaceNorth:
		facing = "north"
	case world.FaceEast:
		facing = "east"
	case world.FaceSouth:
		facing = "south"
	case world.FaceWest:
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

// Hash ...
func (t Torch) Hash() uint64 {
	return hashTorch | (uint64(t.Facing) << 32) | (uint64(t.Type.Uint8()) << 35)
}

// allTorch ...
func allTorch() (torch []world.Block) {
	for i := world.Face(0); i < 6; i++ {
		torch = append(torch, Torch{Type: fire.Normal(), Facing: i})
		torch = append(torch, Torch{Type: fire.Soul(), Facing: i})
	}
	return
}
