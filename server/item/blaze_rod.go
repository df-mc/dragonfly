package item

// BlazeRod is an item exclusively obtained from blazes.
type BlazeRod struct{}

// EncodeItem ...
func (BlazeRod) EncodeItem() (name string, meta int16) {
	return "minecraft:blaze_rod", 0
}
