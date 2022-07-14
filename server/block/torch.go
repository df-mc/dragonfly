package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Torch are non-solid blocks that emit light.
type Torch struct {
	transparent
	empty

	// Facing is the direction from the torch to the block.
	Facing cube.Face
	// Type is the type of fire lighting the torch.
	Type FireType
}

// BreakInfo ...
func (t Torch) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t))
}

// LightEmissionLevel ...
func (t Torch) LightEmissionLevel() uint8 {
	switch t.Type {
	case NormalFire():
		return 14
	default:
		return t.Type.LightLevel()
	}
}

// UseOnBlock ...
func (t Torch) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
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
		w.SetBlock(pos, nil, nil)
	}
}

// HasLiquidDrops ...
func (t Torch) HasLiquidDrops() bool {
	return true
}

// EncodeItem ...
func (t Torch) EncodeItem() (name string, meta int16) {
	switch t.Type {
	case NormalFire():
		return "minecraft:torch", 0
	case SoulFire():
		return "minecraft:soul_torch", 0
	}
	panic("invalid fire type")
}

// EncodeBlock ...
func (t Torch) EncodeBlock() (name string, properties map[string]any) {
	face := t.Facing.String()
	if t.Facing == cube.FaceDown {
		face = "top"
	}
	switch t.Type {
	case NormalFire():
		return "minecraft:torch", map[string]any{"torch_facing_direction": face}
	case SoulFire():
		return "minecraft:soul_torch", map[string]any{"torch_facing_direction": face}
	}
	panic("invalid fire type")
}

// allTorches ...
func allTorches() (torch []world.Block) {
	for i := cube.Face(0); i < 6; i++ {
		if i == cube.FaceUp {
			continue
		}
		torch = append(torch, Torch{Type: NormalFire(), Facing: i})
		torch = append(torch, Torch{Type: SoulFire(), Facing: i})
	}
	return
}
