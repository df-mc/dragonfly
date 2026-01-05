package item

// BreezeRod is a rod dropped by Breeze mobs.
type BreezeRod struct{}

// MaxCount returns the maximum number of Breeze Rods that can be stacked.
func (b BreezeRod) MaxCount() int {
	return 64
}

// EncodeItem encodes the Breeze Rod as an item.
func (b BreezeRod) EncodeItem() (name string, meta int16) {
	return "minecraft:breeze_rod", 0
}