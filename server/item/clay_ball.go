package item

// ClayBall is obtained from mining clay blocks
type ClayBall struct{}

func (ClayBall) SmeltInfo() SmeltInfo {
	return newSmeltInfo(NewStack(Brick{}, 1), 0.3)
}

func (ClayBall) EncodeItem() (name string, meta int16) {
	return "minecraft:clay_ball", 0
}
