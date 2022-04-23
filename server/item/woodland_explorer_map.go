package item

type WoodlandExplorerMap struct {
	FilledMap
}

// EncodeItem ...
func (m WoodlandExplorerMap) EncodeItem() (name string, meta int16) {
	return "minecraft:filled_map", 4
}
