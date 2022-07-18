package item

// RawCopper is a raw metal resource obtained from mining copper ore.
type RawCopper struct{}

// SmeltInfo ...
func (RawCopper) SmeltInfo() SmeltInfo {
	return newOreSmeltInfo(NewStack(CopperIngot{}, 1), 0.7)
}

// EncodeItem ...
func (RawCopper) EncodeItem() (name string, meta int16) {
	return "minecraft:raw_copper", 0
}
