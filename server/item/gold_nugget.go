package item

// GoldNugget is an item used to craft gold ingots & other various gold items.
type GoldNugget struct{}

// EncodeItem ...
func (GoldNugget) EncodeItem() (id int32, name string, meta int16) {
	return 371, "minecraft:gold_nugget", 0
}
