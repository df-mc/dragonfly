package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// CobblestoneStairs are blocks that allow entities to walk up blocks without jumping. They are crafted using Cobblestone.
type CobblestoneStairs struct {
	transparent

	// Mossy specifies if the cobblestone is mossy. This variant of cobblestone is typically found in
	// dungeons or in small clusters in the giant tree taiga biome.
	Mossy bool
	// UpsideDown specifies if the stairs are upside down. If set to true, the full side is at the top part
	// of the block.
	UpsideDown bool
	// Facing is the direction that the full side of the stairs is facing.
	Facing cube.Direction
}

// UseOnBlock handles the directional placing of stairs and makes sure they are properly placed upside down
// when needed.
func (s CobblestoneStairs) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, s)
	if !used {
		return
	}
	s.Facing = user.Facing()
	if face == cube.FaceDown || (clickPos[1] > 0.5 && face != cube.FaceUp) {
		s.UpsideDown = true
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// Model ...
func (s CobblestoneStairs) Model() world.BlockModel {
	return model.Stair{Facing: s.Facing, UpsideDown: s.UpsideDown}
}

// BreakInfo ...
func (s CobblestoneStairs) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(s))
}

// EncodeItem ...
func (s CobblestoneStairs) EncodeItem() (name string, meta int16) {
	if s.Mossy {
		return "minecraft:mossy_cobblestone_stairs", 0
	}
	return "minecraft:stone_stairs", 0
}

// EncodeBlock ...
func (s CobblestoneStairs) EncodeBlock() (name string, properties map[string]any) {
	if s.Mossy {
		return "minecraft:mossy_cobblestone_stairs", map[string]any{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
	}
	return "minecraft:stone_stairs", map[string]any{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
}

// CanDisplace ...
func (CobblestoneStairs) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (s CobblestoneStairs) SideClosed(pos, side cube.Pos, w *world.World) bool {
	return s.Model().FaceSolid(pos, pos.Face(side), w)
}

// allCobblestoneStairs ...
func allCobblestoneStairs() (stairs []world.Block) {
	for direction := cube.Direction(0); direction <= 3; direction++ {
		stairs = append(stairs, CobblestoneStairs{Facing: direction, UpsideDown: true, Mossy: true})
		stairs = append(stairs, CobblestoneStairs{Facing: direction, UpsideDown: false, Mossy: true})
		stairs = append(stairs, CobblestoneStairs{Facing: direction, UpsideDown: true, Mossy: false})
		stairs = append(stairs, CobblestoneStairs{Facing: direction, UpsideDown: false, Mossy: false})
	}
	return
}
