package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// BoneMealResult represents the outcome of a bone meal interaction with a block,
// determining the intensity of the particle effect displayed.
type BoneMealResult int

const (
	// BoneMealResultNone indicates that the bone meal had no effect on the block.
	BoneMealResultNone BoneMealResult = iota
	// BoneMealResultSmall indicates a minor growth effect, produces a small particle burst.
	BoneMealResultSmall
	// BoneMealResultArea indicates a significant growth effect over an area, produces a large particle burst.
	BoneMealResultArea
)

// BoneMeal is an item used to force growth in plants & crops.
type BoneMeal struct{}

// Dispense applies bone meal to the block in front of a dispenser.
func (b BoneMeal) Dispense(pos cube.Pos, face cube.Face, tx *world.Tx, ctx *DispenseContext) DispenseResult {
	front := pos.Side(face)
	affected, ok := tx.Block(front).(BoneMealAffected)
	if !ok {
		return DispenseFailure
	}
	result := affected.BoneMeal(front, tx)
	if result == BoneMealResultNone {
		return DispenseFailure
	}
	ctx.SubtractFromCount(1)
	tx.AddParticle(front.Vec3(), particle.BoneMeal{Area: result == BoneMealResultArea})
	return DispenseSuccess
}

// BoneMealAffected represents a block that is affected when bone meal is used on it.
type BoneMealAffected interface {
	// BoneMeal attempts to affect the block using a bone meal item.
	BoneMeal(pos cube.Pos, tx *world.Tx) BoneMealResult
}

// UseOnBlock ...
func (b BoneMeal) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, tx *world.Tx, _ User, ctx *UseContext) bool {
	if bm, ok := tx.Block(pos).(BoneMealAffected); ok {
		result := bm.BoneMeal(pos, tx)
		if result == BoneMealResultNone {
			return false
		}

		ctx.SubtractFromCount(1)
		tx.AddParticle(pos.Vec3(), particle.BoneMeal{
			Area: result == BoneMealResultArea,
		})
		return true
	}
	return false
}

// EncodeItem ...
func (b BoneMeal) EncodeItem() (name string, meta int16) {
	return "minecraft:bone_meal", 0
}
