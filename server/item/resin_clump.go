package item

// ResinClump ...
type ResinClump struct{}

// SmeltInfo ...
func (ResinClump) SmeltInfo() SmeltInfo {
	return newSmeltInfo(NewStack(ResinBrick{}, 1), 0.3)
}

// EncodeItem ...
func (ResinClump) EncodeItem() (name string, meta int16) {
	return "minecraft:resin_clump", 0
}
