package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

type BoneMealResult int

const (
	BoneMealResultNone BoneMealResult = iota
	BoneMealResultSmall
	BoneMealResultHuge
)

// BoneMeal is an item used to force growth in plants & crops.
type BoneMeal struct{}

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
			Huge: result == BoneMealResultHuge,
		})
		return true
	}
	return false
}

// EncodeItem ...
func (b BoneMeal) EncodeItem() (name string, meta int16) {
	return "minecraft:bone_meal", 0
}
