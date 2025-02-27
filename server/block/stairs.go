package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Stairs are blocks that allow entities to walk up blocks without jumping.
type Stairs struct {
	transparent
	sourceWaterDisplacer

	// Block is the block to use for the type of stair.
	Block world.Block
	// UpsideDown specifies if the stairs are upside down. If set to true, the full side is at the top part
	// of the block.
	UpsideDown bool
	// Facing is the direction that the full side of the stairs is facing.
	Facing cube.Direction
}

// UseOnBlock handles the directional placing of stairs and makes sure they are properly placed upside down
// when needed.
func (s Stairs) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, s)
	if !used {
		return
	}
	s.Facing = user.Rotation().Direction()
	if face == cube.FaceDown || (clickPos[1] > 0.5 && face != cube.FaceUp) {
		s.UpsideDown = true
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// Model ...
func (s Stairs) Model() world.BlockModel {
	return model.Stair{Facing: s.Facing, UpsideDown: s.UpsideDown}
}

// BreakInfo ...
func (s Stairs) BreakInfo() BreakInfo {
	breakInfo := s.Block.(Breakable).BreakInfo()
	return newBreakInfo(breakInfo.Hardness, breakInfo.Harvestable, breakInfo.Effective, oneOf(s)).withBlastResistance(breakInfo.BlastResistance)
}

// Instrument ...
func (s Stairs) Instrument() sound.Instrument {
	if _, ok := s.Block.(Planks); ok {
		return sound.Bass()
	}
	return sound.BassDrum()
}

// FlammabilityInfo ...
func (s Stairs) FlammabilityInfo() FlammabilityInfo {
	if flammable, ok := s.Block.(Flammable); ok {
		return flammable.FlammabilityInfo()
	}
	return newFlammabilityInfo(0, 0, false)
}

// FuelInfo ...
func (s Stairs) FuelInfo() item.FuelInfo {
	if fuel, ok := s.Block.(item.Fuel); ok {
		return fuel.FuelInfo()
	}
	return item.FuelInfo{}
}

// EncodeItem ...
func (s Stairs) EncodeItem() (name string, meta int16) {
	return "minecraft:" + encodeStairsBlock(s.Block) + "_stairs", 0
}

// EncodeBlock ...
func (s Stairs) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + encodeStairsBlock(s.Block) + "_stairs", map[string]any{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
}

// toStairDirection converts a facing to a stair's direction for Minecraft.
func toStairsDirection(v cube.Direction) int32 {
	return int32(3 - v)
}

// SideClosed ...
func (s Stairs) SideClosed(pos, side cube.Pos, tx *world.Tx) bool {
	return s.Model().FaceSolid(pos, pos.Face(side), tx)
}

// allStairs returns all states of stairs.
func allStairs() (stairs []world.Block) {
	f := func(facing cube.Direction, upsideDown bool) {
		for _, s := range StairsBlocks() {
			stairs = append(stairs, Stairs{Facing: facing, UpsideDown: upsideDown, Block: s})
		}
	}
	for i := cube.Direction(0); i <= 3; i++ {
		f(i, true)
		f(i, false)
	}
	return
}
