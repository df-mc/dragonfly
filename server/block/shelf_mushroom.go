package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// ShelfMushroom is a mushroom growing from the side of a block. Entities landing on one are bounced off it.
type ShelfMushroom struct {
	transparent

	// Facing is the direction the shelf mushroom faces, away from the block it is attached to.
	Facing cube.Direction
	// Large is true if the shelf mushroom has grown to its large stage, dropping two rather than one.
	Large bool
}

// Model ...
func (s ShelfMushroom) Model() world.BlockModel {
	return model.ShelfMushroom{Facing: s.Facing, Large: s.Large}
}

// FlammabilityInfo ...
func (ShelfMushroom) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, true)
}

// BreakInfo ...
func (s ShelfMushroom) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, func(item.Tool, []item.Enchantment) []item.Stack {
		if s.Large {
			return []item.Stack{item.NewStack(ShelfMushroom{}, 2)}
		}
		return []item.Stack{item.NewStack(ShelfMushroom{}, 1)}
	})
}

// BoneMeal grows a small shelf mushroom into a large one.
func (s ShelfMushroom) BoneMeal(pos cube.Pos, tx *world.Tx) item.BoneMealResult {
	if s.Large {
		return item.BoneMealResultNone
	}
	s.Large = true
	tx.SetBlock(pos, s, nil)
	return item.BoneMealResultSmall
}

// EntityLand bounces entities off the mushroom, taking part of the fall out of them as a bed does.
func (ShelfMushroom) EntityLand(_ cube.Pos, _ *world.Tx, e world.Entity, distance *float64) {
	if _, ok := e.(fallDistanceEntity); ok {
		*distance *= 0.5
	}
	if v, ok := e.(velocityEntity); ok {
		vel := v.Velocity()
		vel[1] = vel[1] * -3 / 4
		v.SetVelocity(vel)
	}
}

// HasLiquidDrops ...
func (ShelfMushroom) HasLiquidDrops() bool {
	return true
}

// NeighbourUpdateTick ...
func (s ShelfMushroom) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !s.attachable(pos, tx) {
		breakBlock(s, pos, tx)
	}
}

// UseOnBlock ...
func (s ShelfMushroom) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, s)
	if !used || face == cube.FaceUp || face == cube.FaceDown {
		return false
	}
	s.Facing = face.Direction()
	if !s.attachable(pos, tx) {
		return false
	}
	ctx.IgnoreBBox = true

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// attachable returns whether the block behind the mushroom offers it a full side to hang from.
func (s ShelfMushroom) attachable(pos cube.Pos, tx *world.Tx) bool {
	side := pos.Side(s.Facing.Opposite().Face())
	return tx.Block(side).Model().FaceSolid(side, s.Facing.Face(), tx)
}

// EncodeItem ...
func (ShelfMushroom) EncodeItem() (name string, meta int16) {
	return "minecraft:shelf_mushroom", 0
}

// EncodeBlock ...
func (s ShelfMushroom) EncodeBlock() (name string, properties map[string]any) {
	growth := int32(0)
	if s.Large {
		growth = 1
	}
	return "minecraft:shelf_mushroom", map[string]any{
		"minecraft:cardinal_direction": s.Facing.String(),
		"growth":                       growth,
	}
}

// allShelfMushrooms ...
func allShelfMushrooms() (mushrooms []world.Block) {
	for _, d := range cube.Directions() {
		mushrooms = append(mushrooms, ShelfMushroom{Facing: d}, ShelfMushroom{Facing: d, Large: true})
	}
	return
}
