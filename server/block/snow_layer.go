package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
)

// SnowLayer is a ground cover block found on the surface in snowy biomes, and
// can be replenished during snowfall.
type SnowLayer struct {
	gravityAffected
	transparent
	sourceWaterDisplacer

	// Height is the height of the snow layer. It ranges from 0 to 7.
	Height int
	// Covered specifies if the snow layer is covered by another block.
	Covered bool
}

// Model ...
func (s SnowLayer) Model() world.BlockModel {
	return model.SnowLayer{Height: s.Height}
}

// UseOnBlock ...
func (s SnowLayer) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	bottomBlock := tx.Block(pos.Side(cube.FaceDown))
	if _, ok := bottomBlock.(Grass); ok {
		s.Covered = true
	}

	clickedBlock := tx.Block(pos)
	if clickedSnowLayer, ok := clickedBlock.(SnowLayer); ok {
		if clickedSnowLayer.Height < 7 {
			s.Height = clickedSnowLayer.Height + 1
			if s.Height == 7 {
				place(tx, pos, Snow{}, user, ctx)
			} else {
				place(tx, pos, s, user, ctx)
			}
			return placed(ctx)
		}
	}

	pos, _, used = firstReplaceable(tx, pos, face, s)
	if !used {
		return
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (s SnowLayer) BreakInfo() BreakInfo {
	return newBreakInfo(0.1, shovelEffective, shovelEffective, silkTouchDrop(item.NewStack(item.Snowball{}, max(1, s.Height/2+1)), item.NewStack(SnowLayer{}, s.Height+1))).withBlastResistance(0.1)
}

// ScheduledTick ...
func (s SnowLayer) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	s.tick(pos, tx)
}

// RandomTick ...
func (s SnowLayer) RandomTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	s.tick(pos, tx)
}

// tick ...
func (s SnowLayer) tick(pos cube.Pos, tx *world.Tx) {
	bottomBlock := tx.Block(pos.Side(cube.FaceDown))
	if _, ok := bottomBlock.(Grass); ok {
		s.Covered = true
		tx.SetBlock(pos, s, nil)
	}

	if tx.Light(pos) >= 12 {
		newHeight := s.Height - 1
		if newHeight < 0 {
			tx.SetBlock(pos, Air{}, nil)
		} else {
			s.Height = newHeight
			tx.SetBlock(pos, s, nil)
		}
	}
}

// NeighbourUpdateTick ...
func (s SnowLayer) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	s.fall(s, pos, tx)
}

// allSnowLayers ...
func allSnowLayers() (b []world.Block) {
	for i := 0; i <= 7; i++ {
		b = append(b, SnowLayer{Height: i, Covered: false})
		b = append(b, SnowLayer{Height: i, Covered: true})
	}
	return
}

// EncodeItem ...
func (s SnowLayer) EncodeItem() (name string, meta int16) {
	return "minecraft:snow_layer", 0
}

// EncodeBlock ...
func (s SnowLayer) EncodeBlock() (string, map[string]any) {
	return "minecraft:snow_layer", map[string]any{"covered_bit": boolByte(s.Covered), "height": int32(s.Height)}
}
