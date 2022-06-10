package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// WoodStairs are blocks that allow entities to walk up blocks without jumping. They are crafted using planks.
type WoodStairs struct {
	transparent
	bass

	// Wood is the type of wood of the stairs. This field must have one of the values found in the material
	// package.
	Wood WoodType
	// UpsideDown specifies if the stairs are upside down. If set to true, the full side is at the top part
	// of the block.
	UpsideDown bool
	// Facing is the direction that the full side of the stairs is facing.
	Facing cube.Direction
}

// FlammabilityInfo ...
func (s WoodStairs) FlammabilityInfo() FlammabilityInfo {
	if !s.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 20, true)
}

// UseOnBlock handles the directional placing of stairs and makes sure they are properly placed upside down
// when needed.
func (s WoodStairs) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
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
func (s WoodStairs) Model() world.BlockModel {
	return model.Stair{Facing: s.Facing, UpsideDown: s.UpsideDown}
}

// BreakInfo ...
func (s WoodStairs) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(s))
}

// EncodeItem ...
func (s WoodStairs) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Wood.String() + "_stairs", 0
}

// EncodeBlock ...
func (s WoodStairs) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + s.Wood.String() + "_stairs", map[string]any{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
}

// toStairDirection converts a facing to a stair's direction for Minecraft.
func toStairsDirection(v cube.Direction) int32 {
	return int32(3 - v)
}

// CanDisplace ...
func (WoodStairs) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (s WoodStairs) SideClosed(pos, side cube.Pos, w *world.World) bool {
	return s.Model().FaceSolid(pos, pos.Face(side), w)
}

// allWoodStairs returns all states of wood stairs.
func allWoodStairs() (stairs []world.Block) {
	f := func(facing cube.Direction, upsideDown bool) {
		for _, w := range WoodTypes() {
			stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: w})
		}
	}
	for i := cube.Direction(0); i <= 3; i++ {
		f(i, true)
		f(i, false)
	}
	return
}
