package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SnowLayer is a thin, partial covering of snow that may be stacked up to eight layers high.
type SnowLayer struct {
	transparent

	// Height is the number of snow layers in addition to the bottom one, ranging from 0 (a single layer) to 7
	// (eight layers, the height of a full block).
	Height int
}

// Model ...
func (s SnowLayer) Model() world.BlockModel {
	return model.Snow{Layers: s.Height + 1}
}

// ReplaceableBy returns true if more snow may still be stacked onto the layer, or, for any other block, only when
// this is a single layer thin enough to be replaced.
func (s SnowLayer) ReplaceableBy(b world.Block) bool {
	if _, ok := b.(SnowLayer); ok {
		return s.Height < 7
	}
	return s.Height == 0
}

// UseOnBlock places a new snow layer or, when used on an existing layer, stacks another layer onto it.
func (s SnowLayer) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	if existing, ok := tx.Block(pos).(SnowLayer); ok && existing.Height < 7 {
		existing.Height++
		place(tx, pos, existing, user, ctx)
		return placed(ctx)
	}

	pos, _, used := firstReplaceable(tx, pos, face, s)
	if !used {
		return false
	}
	below := pos.Side(cube.FaceDown)
	if !tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx) {
		// Snow can only rest on top of a block with a solid upward face.
		return false
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick breaks the snow layer if the block below it no longer supports it.
func (s SnowLayer) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	below := pos.Side(cube.FaceDown)
	if !tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx) {
		breakBlock(s, pos, tx)
	}
}

// RandomTick melts the snow in biomes too warm to sustain it, removing a single layer at a time until it is gone.
func (s SnowLayer) RandomTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	// TODO: also melt when the block light level is 12 or higher, regardless of the biome temperature. The world
	// does not currently expose the block light level separately from the sky light, which is required to do this
	// without melting snow in daylight.
	if tx.Temperature(pos) <= 0.15 {
		return
	}
	if s.Height == 0 {
		tx.SetBlock(pos, nil, nil)
		return
	}
	s.Height--
	tx.SetBlock(pos, s, nil)
}

// BreakInfo ...
func (s SnowLayer) BreakInfo() BreakInfo {
	return newBreakInfo(0.1, alwaysHarvestable, shovelEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		layers := s.Height + 1
		if hasSilkTouch(enchantments) {
			if layers >= 8 {
				return []item.Stack{item.NewStack(Snow{}, 1)}
			}
			return []item.Stack{item.NewStack(SnowLayer{}, layers)}
		}
		switch {
		case layers <= 3:
			return []item.Stack{item.NewStack(item.Snowball{}, 1)}
		case layers <= 5:
			return []item.Stack{item.NewStack(item.Snowball{}, 2)}
		case layers <= 7:
			return []item.Stack{item.NewStack(item.Snowball{}, 3)}
		default:
			return []item.Stack{item.NewStack(item.Snowball{}, 4)}
		}
	})
}

// EncodeItem ...
func (SnowLayer) EncodeItem() (name string, meta int16) {
	return "minecraft:snow_layer", 0
}

// EncodeBlock ...
func (s SnowLayer) EncodeBlock() (string, map[string]any) {
	return "minecraft:snow_layer", map[string]any{"height": int32(s.Height), "covered_bit": boolByte(false)}
}

// allSnowLayers ...
func allSnowLayers() (s []world.Block) {
	for h := 0; h < 8; h++ {
		s = append(s, SnowLayer{Height: h})
	}
	return
}
