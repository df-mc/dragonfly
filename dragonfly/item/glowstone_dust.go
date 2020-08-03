package item

// GlowstoneDust is dropped when breaking the glowstone block.
type GlowstoneDust struct{}

// EncodeItem ...
func (g GlowstoneDust) EncodeItem() (id int32, meta int16) {
	return 348, 0
}
