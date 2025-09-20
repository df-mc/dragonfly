package item

// PotterySherd is an item that can be found from brushing suspicious sand or gravel.
type PotterySherd struct {
	Type SherdType
}

func (s PotterySherd) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Type.String() + "_pottery_sherd", 0
}

func (PotterySherd) PotDecoration() bool {
	return true
}
