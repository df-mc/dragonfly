package item

// NautilusShell is an item that is used for crafting conduits.
type NautilusShell struct{}

// EncodeItem ...
func (NautilusShell) EncodeItem() (name string, meta int16) {
	return "minecraft:nautilus_shell", 0
}
