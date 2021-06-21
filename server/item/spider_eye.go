package item

// SpiderEye is a poisonous food and brewing item.
type SpiderEye struct{}

// EncodeItem ...
func (SpiderEye) EncodeItem() (name string, meta int16) {
	return "minecraft:spider_eye", 0
}
