package item

// TODO: Documentation
type Diamond struct{}

func (d Diamond) EncodeItem() (id int32, meta int16) {
	return 264, 0
}
