package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Loom is a block used to apply patterns on banners. It is also used as a shepherd's job site block that is found in
// villages.
type Loom struct {
	solid
	bass

	// Facing is the direction the loom is facing.
	Facing cube.Direction
}

func (Loom) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

func (l Loom) BreakInfo() BreakInfo {
	return newBreakInfo(2.5, alwaysHarvestable, axeEffective, oneOf(l))
}

func (Loom) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos, tx)
		return true
	}
	return false
}

func (l Loom) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, l)
	if !used {
		return
	}
	l.Facing = user.Rotation().Direction().Opposite()
	place(tx, pos, l, user, ctx)
	return placed(ctx)
}

func (Loom) EncodeItem() (name string, meta int16) {
	return "minecraft:loom", 0
}

func (l Loom) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:loom", map[string]interface{}{"direction": int32(horizontalDirection(l.Facing))}
}

func allLooms() (looms []world.Block) {
	for _, d := range cube.Directions() {
		looms = append(looms, Loom{Facing: d})
	}
	return
}
