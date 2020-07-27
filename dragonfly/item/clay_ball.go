package item

// ClayBall is obtained from mining clay blocks
type ClayBall struct{}

// EncodeItem ...
func (ClayBall) EncodeItem() (id int32, meta int16) {
	return 337, 0
}
