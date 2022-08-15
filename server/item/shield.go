package item

// Shield is a tool used for protecting the player against attacks.
type Shield struct{}

// MaxCount ...
func (Shield) MaxCount() int {
	return 1
}

// RepairableBy ...
func (Shield) RepairableBy(i Stack) bool {
	if planks, ok := i.Item().(interface{ RepairsWoodTools() bool }); ok {
		return planks.RepairsWoodTools()
	}
	return false
}

// DurabilityInfo ...
func (s Shield) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 336,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// EncodeItem ...
func (Shield) EncodeItem() (name string, meta int16) {
	return "minecraft:shield", 0
}
