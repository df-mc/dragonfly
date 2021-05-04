package item

// GlowstoneDust is dropped when breaking the glowstone block.
type GlowstoneDust struct{}

// EncodeItem ...
func (g GlowstoneDust) EncodeItem() (name string, meta int16) {
	return "minecraft:glowstone_dust", 0
}
