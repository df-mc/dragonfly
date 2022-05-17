package item

// RawGold is a raw metal resource obtained from mining gold ore.
type RawGold struct{}

// SmeltInfo ...
func (RawGold) SmeltInfo() SmeltInfo {
	return SmeltInfo{Product: NewStack(GoldIngot{}, 1), Experience: 1, Ores: true}
}

// EncodeItem ...
func (RawGold) EncodeItem() (name string, meta int16) {
	return "minecraft:raw_gold", 0
}
