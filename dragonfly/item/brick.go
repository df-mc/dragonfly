package item

// Brick - A Common Mineral Resource can be obtain in smelting Clays
// where to get clays?
// Clay blocks can be found near rivers or lakes.Break the clay block by hand or shovel = clay : )

type Brick struct{}

// EncodeItem ...
func (Brick) EncodeItem() (id int32, meta int16) {
	return 337, 0
}
