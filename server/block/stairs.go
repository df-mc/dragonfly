package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Stairs are blocks that allow entities to walk up blocks without jumping. They are crafted using planks.
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
func (s Stairs) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
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
func (s Stairs) Model() world.BlockModel {
	return model.Stair{Facing: s.Facing, UpsideDown: s.UpsideDown}
}

// BreakInfo ...
func (s Stairs) BreakInfo() BreakInfo {
	hardness, blastResistance, harvestable, effective := 2.0, 30.0, pickaxeHarvestable, pickaxeEffective

	switch block := s.Block.(type) {
	// TODO: Copper
	// TODO: Blackstone
	// TODO: Deepslate
	case Planks:
		harvestable = alwaysHarvestable
		effective = axeEffective
		blastResistance = 15.0
	case Prismarine:
		hardness = 1.5
	case Purpur:
		hardness = 1.5
	case Quartz:
		hardness = 0.8
		blastResistance = 4
	case Sandstone:
		if block.Type != SmoothSandstone() {
			hardness = 0.8
			blastResistance = 4
		}
	case Stone:
		hardness = 1.5
	case StoneBricks:
		if block.Type == NormalStoneBricks() {
			hardness = 1.5
		}
	}
	return newBreakInfo(hardness, harvestable, effective, oneOf(s)).withBlastResistance(blastResistance)
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
	if w, ok := s.Block.(Planks); ok && w.Wood.Flammable() {
		return newFlammabilityInfo(5, 20, true)
	}
	return newFlammabilityInfo(0, 0, false)
}

// FuelInfo ...
func (s Stairs) FuelInfo() item.FuelInfo {
	if w, ok := s.Block.(Planks); ok && w.Wood.Flammable() {
		return newFuelInfo(time.Second * 15)
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
func (s Stairs) SideClosed(pos, side cube.Pos, w *world.World) bool {
	return s.Model().FaceSolid(pos, pos.Face(side), w)
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
