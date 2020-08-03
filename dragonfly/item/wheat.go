package item

type Wheat struct{}

func (w Wheat) EncodeItem() (id int32, meta int16) {
	return 296, 0
}
