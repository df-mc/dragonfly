package item

// InkSac is an item dropped by a squid upon death used to create black dye, dark prismarine and book and quill. The
// glowing variant, obtained by killing a glow squid, may be used to cause sign text to light up.
type InkSac struct {
	// Glow specifies if the ink sac is that of a glow squid. If true, it may be used on a sign to light up its text.
	Glow bool
}

// EncodeItem ...
func (i InkSac) EncodeItem() (name string, meta int16) {
	if i.Glow {
		return "minecraft:glow_ink_sac", 0
	}
	return "minecraft:ink_sac", 0
}
