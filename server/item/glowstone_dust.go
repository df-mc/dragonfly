package item

// GlowstoneDust is dropped when breaking the glowstone block.
type GlowstoneDust struct{}

// EncodeItem ...
func (g GlowstoneDust) EncodeItem() (id int32, name string, meta int16) {
	return 348, "minecraft:glowstone_dust", 0
}
