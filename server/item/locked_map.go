package item

type LockedMap struct {
	BaseMap
}

// EncodeItem ...
func (m LockedMap) EncodeItem() (name string, meta int16) {
	return "minecraft:filled_map", 6
}
