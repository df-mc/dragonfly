package item

// GoldNugget is an item used to craft gold ingots & other various gold items.
type GoldNugget struct{}

// EncodeItem ...
func (GoldNugget) EncodeItem() (name string, meta int16) {
	return "minecraft:gold_nugget", 0
}
