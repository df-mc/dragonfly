package item

// ClayBall is obtained from mining clay blocks
type ClayBall struct{}

// SmeltInfo ...
func (ClayBall) SmeltInfo() SmeltInfo {
	return SmeltInfo{
		Product:    NewStack(Brick{}, 1),
		Experience: 0.3,
	}
}

// EncodeItem ...
func (ClayBall) EncodeItem() (name string, meta int16) {
	return "minecraft:clay_ball", 0
}
