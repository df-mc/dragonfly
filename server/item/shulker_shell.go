package item

// ShulkerShell are items dropped by shulkers that are used solely to craft shulker boxes.
type ShulkerShell struct{}

// EncodeItem ...
func (ShulkerShell) EncodeItem() (name string, meta int16) {
	return "minecraft:shulker_shell", 0
}
