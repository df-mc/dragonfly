package item

// GlassBottle is an item that can hold various liquids.
type GlassBottle struct{}

// EncodeItem ...
func (g GlassBottle) EncodeItem() (id int32, meta int16) {
	return 374, 0
}
