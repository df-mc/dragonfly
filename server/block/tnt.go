package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// TNT is a block that blows up blocks around it and activate other TNTs
type TNT struct {
	solid

	Underwater     bool
	IgnitedOnBreak bool
}

// Activate ...
func (t TNT) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User) {
	m, o := u.HeldItems()
	i := item.FlintAndSteel{}
	if m.Item() == i || o.Item() == i {
		w.SetBlock(pos, Air{})
		w.AddEntity(&entity.PrimedTNT{Pos: pos, W: w})
	}
}

// EncodeItem ...
func (TNT) EncodeItem() (name string, meta int16) {
	return "minecraft:tnt", 0
}

// EncodeBlock ...
func (t TNT) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:tnt", map[string]interface{}{
		"allow_underwater_bit": boolByte(t.Underwater),
		"explode_bit":          boolByte(t.IgnitedOnBreak),
	}
}

// BreakInfo ...
func (t TNT) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t))
}

// Explode will cause the TNT to explode and destroy blocks nearby
func (t TNT) Explode() {

}
