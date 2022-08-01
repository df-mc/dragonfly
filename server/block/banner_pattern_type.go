package block

// BannerPatternType represents a type of banner pattern, used to customize banners.
type BannerPatternType struct {
	bannerPatternType
}

// BorderBannerPattern represents the 'Border' banner pattern type.
func BorderBannerPattern() BannerPatternType {
	return BannerPatternType{0}
}

// BricksBannerPattern represents the 'Bricks' banner pattern type.
func BricksBannerPattern() BannerPatternType {
	return BannerPatternType{1}
}

// CircleBannerPattern represents the 'Circle' banner pattern type.
func CircleBannerPattern() BannerPatternType {
	return BannerPatternType{2}
}

// CreeperBannerPattern represents the 'Creeper' banner pattern type.
func CreeperBannerPattern() BannerPatternType {
	return BannerPatternType{3}
}

// CrossBannerPattern represents the 'Cross' banner pattern type.
func CrossBannerPattern() BannerPatternType {
	return BannerPatternType{4}
}

// CurlyBorderBannerPattern represents the 'Curly Border' banner pattern type.
func CurlyBorderBannerPattern() BannerPatternType {
	return BannerPatternType{5}
}

// DiagonalLeftBannerPattern represents the 'Diagonal Left' banner pattern type.
func DiagonalLeftBannerPattern() BannerPatternType {
	return BannerPatternType{6}
}

// DiagonalRightBannerPattern represents the 'Diagonal Right' banner pattern type.
func DiagonalRightBannerPattern() BannerPatternType {
	return BannerPatternType{7}
}

// DiagonalUpLeftBannerPattern represents the 'Diagonal Up Left' banner pattern type.
func DiagonalUpLeftBannerPattern() BannerPatternType {
	return BannerPatternType{8}
}

// DiagonalUpRightBannerPattern represents the 'Diagonal Up Right' banner pattern type.
func DiagonalUpRightBannerPattern() BannerPatternType {
	return BannerPatternType{9}
}

// FlowerBannerPattern represents the 'Flower' banner pattern type.
func FlowerBannerPattern() BannerPatternType {
	return BannerPatternType{10}
}

// GradientBannerPattern represents the 'Gradient' banner pattern type.
func GradientBannerPattern() BannerPatternType {
	return BannerPatternType{11}
}

// GradientUpBannerPattern represents the 'Gradient Up' banner pattern type.
func GradientUpBannerPattern() BannerPatternType {
	return BannerPatternType{12}
}

// HalfHorizontalBannerPattern represents the 'Half Horizontal' banner pattern type.
func HalfHorizontalBannerPattern() BannerPatternType {
	return BannerPatternType{13}
}

// HalfHorizontalBottomBannerPattern represents the 'Half Horizontal Bottom' banner pattern type.
func HalfHorizontalBottomBannerPattern() BannerPatternType {
	return BannerPatternType{14}
}

// HalfVerticalBannerPattern represents the 'Half Vertical' banner pattern type.
func HalfVerticalBannerPattern() BannerPatternType {
	return BannerPatternType{15}
}

// HalfVerticalRightBannerPattern represents the 'Half Vertical Right' banner pattern type.
func HalfVerticalRightBannerPattern() BannerPatternType {
	return BannerPatternType{16}
}

// MojangBannerPattern represents the 'Mojang' banner pattern type.
func MojangBannerPattern() BannerPatternType {
	return BannerPatternType{17}
}

// RhombusBannerPattern represents the 'Rhombus' banner pattern type.
func RhombusBannerPattern() BannerPatternType {
	return BannerPatternType{18}
}

// SkullBannerPattern represents the 'Skull' banner pattern type.
func SkullBannerPattern() BannerPatternType {
	return BannerPatternType{19}
}

// SmallStripesBannerPattern represents the 'Small Stripes' banner pattern type.
func SmallStripesBannerPattern() BannerPatternType {
	return BannerPatternType{20}
}

// SquareBottomLeftBannerPattern represents the 'Square Bottom Left' banner pattern type.
func SquareBottomLeftBannerPattern() BannerPatternType {
	return BannerPatternType{21}
}

// SquareBottomRightBannerPattern represents the 'Square Bottom Right' banner pattern type.
func SquareBottomRightBannerPattern() BannerPatternType {
	return BannerPatternType{22}
}

// SquareTopLeftBannerPattern represents the 'Square Top Left' banner pattern type.
func SquareTopLeftBannerPattern() BannerPatternType {
	return BannerPatternType{23}
}

// SquareTopRightBannerPattern represents the 'Square Top Right' banner pattern type.
func SquareTopRightBannerPattern() BannerPatternType {
	return BannerPatternType{24}
}

// StraightCrossBannerPattern represents the 'Straight Cross' banner pattern type.
func StraightCrossBannerPattern() BannerPatternType {
	return BannerPatternType{25}
}

// StripeBottomBannerPattern represents the 'Stripe Bottom' banner pattern type.
func StripeBottomBannerPattern() BannerPatternType {
	return BannerPatternType{26}
}

// StripeCenterBannerPattern represents the 'Stripe Center' banner pattern type.
func StripeCenterBannerPattern() BannerPatternType {
	return BannerPatternType{27}
}

// StripeDownLeftBannerPattern represents the 'Stripe Down Left' banner pattern type.
func StripeDownLeftBannerPattern() BannerPatternType {
	return BannerPatternType{28}
}

// StripeDownRightBannerPattern represents the 'Stripe Down Right' banner pattern type.
func StripeDownRightBannerPattern() BannerPatternType {
	return BannerPatternType{29}
}

// StripeLeftBannerPattern represents the 'Stripe Left' banner pattern type.
func StripeLeftBannerPattern() BannerPatternType {
	return BannerPatternType{30}
}

// StripeMiddleBannerPattern represents the 'Stripe Middle' banner pattern type.
func StripeMiddleBannerPattern() BannerPatternType {
	return BannerPatternType{31}
}

// StripeRightBannerPattern represents the 'Stripe Right' banner pattern type.
func StripeRightBannerPattern() BannerPatternType {
	return BannerPatternType{32}
}

// StripeTopBannerPattern represents the 'Stripe Top' banner pattern type.
func StripeTopBannerPattern() BannerPatternType {
	return BannerPatternType{33}
}

// TriangleBottomBannerPattern represents the 'Triangle Bottom' banner pattern type.
func TriangleBottomBannerPattern() BannerPatternType {
	return BannerPatternType{34}
}

// TriangleTopBannerPattern represents the 'Triangle Top' banner pattern type.
func TriangleTopBannerPattern() BannerPatternType {
	return BannerPatternType{35}
}

// TrianglesBottomBannerPattern represents the 'Triangles Bottom' banner pattern type.
func TrianglesBottomBannerPattern() BannerPatternType {
	return BannerPatternType{36}
}

// TrianglesTopBannerPattern represents the 'Triangles Top' banner pattern type.
func TrianglesTopBannerPattern() BannerPatternType {
	return BannerPatternType{37}
}

// GlobeBannerPattern represents the 'Globe' banner pattern type.
func GlobeBannerPattern() BannerPatternType {
	return BannerPatternType{38}
}

// PiglinBannerPattern represents the 'Piglin' banner pattern type.
func PiglinBannerPattern() BannerPatternType {
	return BannerPatternType{39}
}

// BannerPatternTypes returns all the available banner pattern types.
func BannerPatternTypes() []BannerPatternType {
	return []BannerPatternType{
		BorderBannerPattern(),
		BricksBannerPattern(),
		CircleBannerPattern(),
		CreeperBannerPattern(),
		CrossBannerPattern(),
		CurlyBorderBannerPattern(),
		DiagonalLeftBannerPattern(),
		DiagonalRightBannerPattern(),
		DiagonalUpLeftBannerPattern(),
		DiagonalUpRightBannerPattern(),
		FlowerBannerPattern(),
		GradientBannerPattern(),
		GradientUpBannerPattern(),
		HalfHorizontalBannerPattern(),
		HalfHorizontalBottomBannerPattern(),
		HalfVerticalBannerPattern(),
		HalfVerticalRightBannerPattern(),
		MojangBannerPattern(),
		RhombusBannerPattern(),
		SkullBannerPattern(),
		SmallStripesBannerPattern(),
		SquareBottomLeftBannerPattern(),
		SquareBottomRightBannerPattern(),
		SquareTopLeftBannerPattern(),
		SquareTopRightBannerPattern(),
		StraightCrossBannerPattern(),
		StripeBottomBannerPattern(),
		StripeCenterBannerPattern(),
		StripeDownLeftBannerPattern(),
		StripeDownRightBannerPattern(),
		StripeLeftBannerPattern(),
		StripeMiddleBannerPattern(),
		StripeRightBannerPattern(),
		StripeTopBannerPattern(),
		TriangleBottomBannerPattern(),
		TriangleTopBannerPattern(),
		TrianglesBottomBannerPattern(),
		TrianglesTopBannerPattern(),
		GlobeBannerPattern(),
		PiglinBannerPattern(),
	}
}

type bannerPatternType uint8

// Uint8 returns the bannerPatternType as a uint8.
func (b bannerPatternType) Uint8() uint8 {
	return uint8(b)
}

// String returns the bannerPatternType as a string.
func (b bannerPatternType) String() string {
	switch b {
	case 0:
		return "border"
	case 1:
		return "bricks"
	case 2:
		return "circle"
	case 3:
		return "creeper"
	case 4:
		return "cross"
	case 5:
		return "curly_border"
	case 6:
		return "diagonal_left"
	case 7:
		return "diagonal_right"
	case 8:
		return "diagonal_up_left"
	case 9:
		return "diagonal_up_right"
	case 10:
		return "flower"
	case 11:
		return "gradient"
	case 12:
		return "gradient_up"
	case 13:
		return "half_horizontal"
	case 14:
		return "half_horizontal_bottom"
	case 15:
		return "half_vertical"
	case 16:
		return "half_vertical_right"
	case 17:
		return "mojang"
	case 18:
		return "rhombus"
	case 19:
		return "skull"
	case 20:
		return "small_stripes"
	case 21:
		return "square_bottom_left"
	case 22:
		return "square_bottom_right"
	case 23:
		return "square_top_left"
	case 24:
		return "square_top_right"
	case 25:
		return "straight_cross"
	case 26:
		return "stripe_bottom"
	case 27:
		return "stripe_center"
	case 28:
		return "stripe_downleft"
	case 29:
		return "stripe_downright"
	case 30:
		return "stripe_left"
	case 31:
		return "stripe_middle"
	case 32:
		return "stripe_right"
	case 33:
		return "stripe_top"
	case 34:
		return "triangle_bottom"
	case 35:
		return "triangle_top"
	case 36:
		return "triangles_bottom"
	case 37:
		return "triangles_top"
	case 38:
		return "globe"
	case 39:
		return "piglin"
	}
	panic("should never happen")
}
