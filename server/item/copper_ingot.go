package item

// CopperIngot is a metal ingot melted from copper ore.
type CopperIngot struct{}

// EncodeItem ...
func (c CopperIngot) EncodeItem() (name string, meta int16) {
	return "minecraft:copper_ingot", 0
}
