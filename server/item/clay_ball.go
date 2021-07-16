package item

// ClayBall is obtained from mining clay blocks
type ClayBall struct{}

// EncodeItem ...
func (ClayBall) EncodeItem() (name string, meta int16) {
	return "minecraft:clay_ball", 0
}
