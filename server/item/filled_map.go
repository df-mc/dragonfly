package item

type FilledMap struct {
	BaseMap
}

// EncodeItem ...
func (m FilledMap) EncodeItem() (name string, meta int16) {
	return "minecraft:filled_map", 0
}
