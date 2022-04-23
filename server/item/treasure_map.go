package item

type TreasureMap struct {
	FilledMap
}

// EncodeItem ...
func (m TreasureMap) EncodeItem() (name string, meta int16) {
	return "minecraft:filled_map", 5
}
