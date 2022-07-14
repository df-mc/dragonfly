package block

import "github.com/df-mc/dragonfly/server/item"

// Bookshelf is a decorative block that primarily serves to enhance enchanting with an enchanting table.
type Bookshelf struct {
	solid
	bass
}

// BreakInfo ...
func (b Bookshelf) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, alwaysHarvestable, axeEffective, silkTouchDrop(item.NewStack(item.Book{}, 3), item.NewStack(b, 1)))
}

// EncodeItem ...
func (Bookshelf) EncodeItem() (name string, meta int16) {
	return "minecraft:bookshelf", 0
}

// EncodeBlock ...
func (Bookshelf) EncodeBlock() (string, map[string]any) {
	return "minecraft:bookshelf", nil
}
