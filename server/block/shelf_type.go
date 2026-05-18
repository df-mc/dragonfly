package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Shelf is a wooden block that can be used for decoration. It supports all wood types found in Dragonfly.
type Shelf struct {
	solid
	bass
	// Wood is the type of wood of the shelf.
	Wood WoodType
	// Facing is the direction the shelf is facing.
	Facing cube.Direction
}

// MaxCount ...
func (Shelf) MaxCount() int {
	return 64
}

// FlammabilityInfo ...
func (s Shelf) FlammabilityInfo() FlammabilityInfo {
	if !s.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 20, true)
}

// FuelInfo ...
func (s Shelf) FuelInfo() item.FuelInfo {
	if !s.Wood.Flammable() {
		return item.FuelInfo{}
	}
	return newFuelInfo(time.Second * 15)
}

// EncodeItem ...
func (s Shelf) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Wood.String() + "_shelf", 0
}

// BreakInfo ...
func (s Shelf) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(s)).withBlastResistance(15)
}

// UseOnBlock ...
func (s Shelf) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, s)
	if !used {
		return false
	}
	s.Facing = user.Rotation().Direction()
	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// EncodeBlock ...
func (s Shelf) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + s.Wood.String() + "_shelf", map[string]any{
		"facing_direction": int32(s.Facing.Face()),
	}
}

// allShelves returns a list of all shelf block states.
func allShelves() (shelves []world.Block) {
	for _, w := range WoodTypes() {
		for _, d := range cube.Directions() {
			shelves = append(shelves, Shelf{Wood: w, Facing: d})
		}
	}
	return
}
