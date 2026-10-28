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
	// Corner is the corner that the stairs form with the stairs around them.
	Corner StairsCorner
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

	s.Corner = s.calculateCorner(tx, pos)
	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (s Stairs) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if corner := s.calculateCorner(tx, pos); corner != s.Corner {
		s.Corner = corner
		tx.SetBlock(pos, s, nil)
	}
}

// DeriveState ...
func (s Stairs) DeriveState(pos cube.Pos, src world.BlockSource) world.Block {
	s.Corner = s.calculateCorner(src, pos)
	return s
}

// calculateCorner calculates the corner that the stairs form with the stairs around them at a position in a world.
func (s Stairs) calculateCorner(src world.BlockSource, pos cube.Pos) StairsCorner {
	left, right := s.Facing.RotateLeft(), s.Facing.RotateRight()
	if front, ok := src.Block(pos.Side(s.Facing.Face())).(Stairs); ok && front.UpsideDown == s.UpsideDown {
		if front.Facing == left && !s.continued(src, pos, right) {
			return OuterLeftStairsCorner()
		} else if front.Facing == right && !s.continued(src, pos, left) {
			return OuterRightStairsCorner()
		}
	}
	if back, ok := src.Block(pos.Side(s.Facing.Opposite().Face())).(Stairs); ok && back.UpsideDown == s.UpsideDown {
		if back.Facing == left && !s.continued(src, pos, left) {
			return InnerLeftStairsCorner()
		} else if back.Facing == right && !s.continued(src, pos, right) {
			return InnerRightStairsCorner()
		}
	}
	return NoStairsCorner()
}

// continued returns true if the stairs are continued in the direction passed by stairs of the same type facing the
// same way. Stairs like these do not form a corner on that side.
func (s Stairs) continued(src world.BlockSource, pos cube.Pos, d cube.Direction) bool {
	side, ok := src.Block(pos.Side(d.Face())).(Stairs)
	return ok && side.Facing == s.Facing && side.UpsideDown == s.UpsideDown && encodeStairsBlock(side.Block) == encodeStairsBlock(s.Block)
}

// Model ...
func (s Stairs) Model() world.BlockModel {
	return model.Stair{
		Facing:     s.Facing,
		UpsideDown: s.UpsideDown,
		Corner:     s.Corner != NoStairsCorner(),
		Inner:      s.Corner.Inner(),
		Left:       s.Corner.Left(),
	}
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
	if _, ok := s.Block.(BambooMosaic); ok {
		return sound.Bass()
	}
	if _, ok := s.Block.(Wool); ok {
		return sound.Guitar()
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
	return "minecraft:" + encodeStairsBlock(s.Block) + "_stairs", map[string]any{"minecraft:corner": s.Corner.String(), "upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
}

// toStairDirection converts a facing to a stair's direction for Minecraft.
func toStairsDirection(v cube.Direction) int32 {
	return int32(3 - v)
}

// SideClosed ...
func (s Stairs) SideClosed(pos, side cube.Pos, tx *world.Tx) bool {
	return s.Model().FaceSolid(pos, pos.Face(side), tx)
}

// CanRedstoneWireStepDown ...
func (Stairs) CanRedstoneWireStepDown(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// allStairs returns all states of stairs.
func allStairs() (stairs []world.Block) {
	f := func(facing cube.Direction, upsideDown bool) {
		for _, s := range StairsBlocks() {
			for _, corner := range StairsCorners() {
				stairs = append(stairs, Stairs{Facing: facing, UpsideDown: upsideDown, Block: s, Corner: corner})
			}
		}
	}
	for i := cube.Direction(0); i <= 3; i++ {
		f(i, true)
		f(i, false)
	}
	return
}
