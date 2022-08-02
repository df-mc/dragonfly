package item

// BannerPatternType represents a type of BannerPattern.
type BannerPatternType struct {
	bannerPatternType
}

// CreeperBannerPattern represents the 'Creeper' banner pattern type.
func CreeperBannerPattern() BannerPatternType {
	return BannerPatternType{0}
}

// SkullBannerPattern represents the 'Skull' banner pattern type.
func SkullBannerPattern() BannerPatternType {
	return BannerPatternType{1}
}

// FlowerBannerPattern represents the 'Flower' banner pattern type.
func FlowerBannerPattern() BannerPatternType {
	return BannerPatternType{2}
}

// MojangBannerPattern represents the 'Mojang' banner pattern type.
func MojangBannerPattern() BannerPatternType {
	return BannerPatternType{3}
}

// FieldMasonedBannerPattern represents the 'Field Masoned' banner pattern type.
func FieldMasonedBannerPattern() BannerPatternType {
	return BannerPatternType{4}
}

// BordureIndentedBannerPattern represents the 'Bordure Indented' banner pattern type.
func BordureIndentedBannerPattern() BannerPatternType {
	return BannerPatternType{5}
}

// PiglinBannerPattern represents the 'Piglin' banner pattern type.
func PiglinBannerPattern() BannerPatternType {
	return BannerPatternType{6}
}

// GlobeBannerPattern represents the 'Globe' banner pattern type.
func GlobeBannerPattern() BannerPatternType {
	return BannerPatternType{7}
}

// BannerPatterns returns all possible banner patterns.
func BannerPatterns() []BannerPatternType {
	return []BannerPatternType{
		CreeperBannerPattern(),
		SkullBannerPattern(),
		FlowerBannerPattern(),
		MojangBannerPattern(),
		FieldMasonedBannerPattern(),
		BordureIndentedBannerPattern(),
		PiglinBannerPattern(),
		GlobeBannerPattern(),
	}
}

type bannerPatternType uint8

// Uint8 returns the uint8 value of the banner pattern type.
func (b bannerPatternType) Uint8() uint8 {
	return uint8(b)
}

// String ...
func (b bannerPatternType) String() string {
	switch b {
	case 0:
		return "creeper"
	case 1:
		return "skull"
	case 2:
		return "flower"
	case 3:
		return "mojang"
	case 4:
		return "field_masoned"
	case 5:
		return "bordure_indented"
	case 6:
		return "piglin"
	case 7:
		return "globe"
	}
	panic("should never happen")
}
