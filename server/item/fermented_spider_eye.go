package item

// FermentedSpiderEye is a brewing ingredient.
type FermentedSpiderEye struct{}

// EncodeItem ...
func (FermentedSpiderEye) EncodeItem() (name string, meta int16) {
	return "minecraft:fermented_spider_eye", 0
}
