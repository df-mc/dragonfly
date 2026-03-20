package block

import (
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SweetBerries are edible berries that can also be planted to grow a sweet berry bush.
type SweetBerries struct{}

// AlwaysConsumable ...
func (SweetBerries) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (SweetBerries) ConsumeDuration() time.Duration {
	return item.DefaultConsumeDuration
}

// Consume ...
func (SweetBerries) Consume(_ *world.Tx, c item.Consumer) item.Stack {
	c.Saturate(2, 1.2)
	return item.Stack{}
}

// CompostChance ...
func (SweetBerries) CompostChance() float64 {
	return 0.3
}

// UseOnBlock ...
func (s SweetBerries) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	bush := SweetBerryBush{}
	pos, _, used := firstReplaceable(tx, pos, face, bush)
	if !used || !supportsVegetation(bush, tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	place(tx, pos, bush, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (SweetBerries) EncodeItem() (name string, meta int16) {
	return "minecraft:sweet_berries", 0
}
