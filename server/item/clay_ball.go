package item

// ClayBall is obtained from mining clay blocks
type ClayBall struct{}

// EncodeItem ...
func (ClayBall) EncodeItem() (id int32, name string, meta int16) {
	return 337, "minecraft:clay_ball", 0
}
