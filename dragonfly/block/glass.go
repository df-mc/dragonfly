package block

import "github.com/df-mc/dragonfly/dragonfly/item"

type Glass struct{}

func (g Glass) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 0.3,
		Drops:    simpleDrops(item.NewStack(g, 1)),
	}
}

// EncodeItem ...
func (g Glass) EncodeItem() (id int32, meta int16) {
	return 20, 0
}

// EncodeBlock ...
func (g Glass) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:glass", nil
}
