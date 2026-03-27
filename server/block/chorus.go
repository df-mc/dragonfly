package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// ChorusPlant is a branching End plant.
type ChorusPlant struct {
	transparent
}

// Model ...
func (ChorusPlant) Model() world.BlockModel {
	return model.Chorus{}
}

// BreakInfo ...
func (c ChorusPlant) BreakInfo() BreakInfo {
	return newBreakInfo(0.4, alwaysHarvestable, nothingEffective, oneOf(c))
}

// NeighbourUpdateTick ...
func (c ChorusPlant) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsChorus(tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(c, pos, tx)
	}
}

// UseOnBlock ...
func (c ChorusPlant) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, c)
	if !used || !supportsChorus(tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}
	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (ChorusPlant) EncodeItem() (name string, meta int16) {
	return "minecraft:chorus_plant", 0
}

// EncodeBlock ...
func (ChorusPlant) EncodeBlock() (string, map[string]any) {
	return "minecraft:chorus_plant", nil
}

// ChorusFlower is the flower that grows on chorus plants.
type ChorusFlower struct {
	transparent

	Age int
}

// Model ...
func (ChorusFlower) Model() world.BlockModel {
	return model.Chorus{}
}

// BreakInfo ...
func (c ChorusFlower) BreakInfo() BreakInfo {
	return newBreakInfo(0.4, alwaysHarvestable, nothingEffective, oneOf(c))
}

// NeighbourUpdateTick ...
func (c ChorusFlower) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsChorus(tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(c, pos, tx)
	}
}

// UseOnBlock ...
func (c ChorusFlower) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, c)
	if !used || !supportsChorus(tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}
	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (ChorusFlower) EncodeItem() (name string, meta int16) {
	return "minecraft:chorus_flower", 0
}

// EncodeBlock ...
func (c ChorusFlower) EncodeBlock() (string, map[string]any) {
	return "minecraft:chorus_flower", map[string]any{"age": int32(c.Age)}
}

func supportsChorus(b world.Block) bool {
	switch b.(type) {
	case EndStone, ChorusPlant:
		return true
	default:
		return false
	}
}
