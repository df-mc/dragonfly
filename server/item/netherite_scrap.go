package item

// NetheriteScrap is a material smelted from ancient debris, which is found in the Nether.
type NetheriteScrap struct{}

// EncodeItem ...
func (NetheriteScrap) EncodeItem() (name string, meta int16) {
	return "minecraft:netherite_scrap", 0
}
