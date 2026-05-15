package item

// Shield is a defensive item that can block incoming attacks while held.
type Shield struct{}

// DurabilityInfo ...
func (Shield) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 337,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// RepairableBy ...
func (Shield) RepairableBy(i Stack) bool {
	return toolTierRepairable(ToolTierWood)(i)
}

// MaxCount always returns 1.
func (Shield) MaxCount() int {
	return 1
}

// OffHand ...
func (Shield) OffHand() bool {
	return true
}

// EncodeItem ...
func (Shield) EncodeItem() (name string, meta int16) {
	return "minecraft:shield", 0
}
