package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// String is an item dropped by spiders and cobwebs. When placed, it turns into a tripwire.
type String struct {
	transparent
}

// UseOnBlock ...
func (s String) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, s)
	if !used {
		return
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (s String) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(String{}))
}

func (s String) Model() world.BlockModel {
	return model.Empty{}
}

// EncodeItem ...
func (String) EncodeItem() (name string, meta int16) {
	return "minecraft:string", 0
}

// EncodeBlock ...
func (s String) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:tripWire", map[string]interface{}{"attached_bit": uint8(0), "disarmed_bit": uint8(0), "powered_bit": uint8(0), "suspended_bit": uint8(1)}
}
