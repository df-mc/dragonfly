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
func (t Torch) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, t)
	if !used {
		return false
	}
	if face == cube.FaceDown {
		return false
	}
	if !tx.Block(pos.Side(face.Opposite())).Model().FaceSolid(pos.Side(face.Opposite()), face, tx) {
		found := false
		for _, i := range []cube.Face{cube.FaceSouth, cube.FaceWest, cube.FaceNorth, cube.FaceEast, cube.FaceDown} {
			if tx.Block(pos.Side(i)).Model().FaceSolid(pos.Side(i), i.Opposite(), tx) {
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

	place(tx, pos, t, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (t Torch) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), tx) {
		breakBlock(t, pos, tx)
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
	var face string
	if t.Facing == cube.FaceDown {
		face = "top"
	} else if t.Facing == unknownFace {
		face = "unknown"
	} else {
		face = t.Facing.String()
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
	for _, face := range cube.Faces() {
		if face == cube.FaceUp {
			face = unknownFace
		}
		torch = append(torch, Torch{Type: NormalFire(), Facing: face})
		torch = append(torch, Torch{Type: SoulFire(), Facing: face})
	}
	return
}
