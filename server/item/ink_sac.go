package item

// An ink sac is an item dropped by a squid upon death used to create black dye, dark prismarine and book and quill.
type InkSac struct{}

// EncodeItem ...
func (InkSac) EncodeItem() (name string, meta int16) {
	return "minecraft:ink_sac", 0
}
