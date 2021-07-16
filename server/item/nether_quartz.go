package item

// NetherQuartz is a smooth, white mineral found in the Nether.
type NetherQuartz struct{}

// EncodeItem ...
func (NetherQuartz) EncodeItem() (name string, meta int16) {
	return "minecraft:quartz", 0
}
