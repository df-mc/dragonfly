package item

// NautilusShell is an item that is used for crafting conduits.
type NautilusShell struct{}

func (NautilusShell) EncodeItem() (name string, meta int16) {
	return "minecraft:nautilus_shell", 0
}

func (NautilusShell) OffHand() bool {
	return true
}
