package item

// GlowInkSac is an item dropped by a glow squid upon death.
type GlowInkSac struct{}

// EncodeItem ...
func (GlowInkSac) EncodeItem() (name string, meta int16) {
	return "minecraft:glow_ink_sac", 0
}
