package block

// Air is the block present in otherwise empty space.
type Air struct {
	empty
	replaceable
	transparent
}

func (Air) HasLiquidDrops() bool {
	return false
}

func (Air) EncodeItem() (name string, meta int16) {
	return "minecraft:air", 0
}

func (Air) EncodeBlock() (string, map[string]any) {
	return "minecraft:air", nil
}
