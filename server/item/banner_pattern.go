package item

// BannerPattern is an item used to customize banners inside looms.
type BannerPattern struct {
	// Type represents the type of banner pattern. These types do not include all patterns that can be applied to a
	// banner.
	Type BannerPatternType
}

// MaxCount always returns 1.
func (b BannerPattern) MaxCount() int {
	return 1
}

// EncodeItem ...
func (b BannerPattern) EncodeItem() (name string, meta int16) {
	return "minecraft:" + b.Type.String() + "_banner_pattern", 0
}
