package block

import "github.com/df-mc/dragonfly/server/item"

// Snow is a full-sized block of snow.
type Snow struct {
	solid
}

// BreakInfo ...
func (s Snow) BreakInfo() BreakInfo {
	return NewBreakInfo(0.2, AlwaysHarvestable, ShovelEffective, SilkTouchDrop(item.NewStack(item.Snowball{}, 4), item.NewStack(s, 1)))
}

// EncodeItem ...
func (Snow) EncodeItem() (name string, meta int16) {
	return "minecraft:snow", 0
}

// EncodeBlock ...
func (Snow) EncodeBlock() (string, map[string]any) {
	return "minecraft:snow", nil
}
