package item

// FermentedSpiderEye is a brewing ingredient.
type FermentedSpiderEye struct{}

func (FermentedSpiderEye) EncodeItem() (name string, meta int16) {
	return "minecraft:fermented_spider_eye", 0
}
