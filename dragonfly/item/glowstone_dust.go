package item

type GlowstoneDust struct {}

func (g GlowstoneDust) EncodeItem() (id int32, meta int16) {
	return 348, 0
}

func (g GlowstoneDust) MaxCount() int {
	return 64
}