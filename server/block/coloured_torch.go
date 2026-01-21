package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// ColouredTorch are non-solid blocks that emit light.
type ColouredTorch struct {
	transparent
	empty

	// Facing is the direction from the torch to the block.
	Facing cube.Face
	// Colour is the colour of the torch.
	Colour TorchColour
}

// BreakInfo ...
func (t ColouredTorch) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t))
}

// LightEmissionLevel ...
func (t ColouredTorch) LightEmissionLevel() uint8 {
	return 14
}

// UseOnBlock ...
func (t ColouredTorch) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
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
func (t ColouredTorch) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), tx) {
		breakBlock(t, pos, tx)
	}
}

// HasLiquidDrops ...
func (t ColouredTorch) HasLiquidDrops() bool {
	return true
}

// EncodeItem ...
func (t ColouredTorch) EncodeItem() (name string, meta int16) {
	return "minecraft:colored_" + t.Colour.String() + "_torch", 0
}

// EncodeBlock ...
func (t ColouredTorch) EncodeBlock() (name string, properties map[string]any) {
	var face string
	if t.Facing == cube.FaceDown {
		face = "top"
	} else if t.Facing == unknownFace {
		face = "unknown"
	} else {
		face = t.Facing.String()
	}

	return "minecraft:colored_" + t.Colour.String() + "_torch", map[string]any{"torch_facing_direction": face}
}

// allColouredTorches ...
func allColouredTorches() (torch []world.Block) {
	for _, face := range cube.Faces() {
		if face == cube.FaceUp {
			face = unknownFace
		}

		for _, colour := range TorchColours() {
			torch = append(torch, ColouredTorch{Facing: face, Colour: colour})
		}
	}
	return
}
