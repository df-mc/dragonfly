package block

var (
	bannerPatternsMap = map[string]BannerPatternType{}
	bannerPatternIDs  = map[BannerPatternType]string{}
)

// init initializes all default banner patterns to the registry.
func init() {
	registerBannerPattern("bo", BorderBannerPattern())
	registerBannerPattern("bri", BricksBannerPattern())
	registerBannerPattern("mc", CircleBannerPattern())
	registerBannerPattern("cre", CreeperBannerPattern())
	registerBannerPattern("cr", CrossBannerPattern())
	registerBannerPattern("cbo", CurlyBorderBannerPattern())
	registerBannerPattern("lud", DiagonalLeftBannerPattern())
	registerBannerPattern("rd", DiagonalRightBannerPattern())
	registerBannerPattern("ld", DiagonalUpLeftBannerPattern())
	registerBannerPattern("rud", DiagonalUpRightBannerPattern())
	registerBannerPattern("flo", FlowerBannerPattern())
	registerBannerPattern("gra", GradientBannerPattern())
	registerBannerPattern("gru", GradientUpBannerPattern())
	registerBannerPattern("hh", HalfHorizontalBannerPattern())
	registerBannerPattern("hhb", HalfHorizontalBottomBannerPattern())
	registerBannerPattern("vh", HalfVerticalBannerPattern())
	registerBannerPattern("vhr", HalfVerticalRightBannerPattern())
	registerBannerPattern("moj", MojangBannerPattern())
	registerBannerPattern("mr", RhombusBannerPattern())
	registerBannerPattern("sku", SkullBannerPattern())
	registerBannerPattern("ss", SmallStripesBannerPattern())
	registerBannerPattern("bl", SquareBottomLeftBannerPattern())
	registerBannerPattern("br", SquareBottomRightBannerPattern())
	registerBannerPattern("tl", SquareTopLeftBannerPattern())
	registerBannerPattern("tr", SquareTopRightBannerPattern())
	registerBannerPattern("sc", StraightCrossBannerPattern())
	registerBannerPattern("bs", StripeBottomBannerPattern())
	registerBannerPattern("cs", StripeCenterBannerPattern())
	registerBannerPattern("dls", StripeDownLeftBannerPattern())
	registerBannerPattern("drs", StripeDownRightBannerPattern())
	registerBannerPattern("ls", StripeLeftBannerPattern())
	registerBannerPattern("ms", StripeMiddleBannerPattern())
	registerBannerPattern("rs", StripeRightBannerPattern())
	registerBannerPattern("ts", StripeTopBannerPattern())
	registerBannerPattern("bt", TriangleBottomBannerPattern())
	registerBannerPattern("tt", TriangleTopBannerPattern())
	registerBannerPattern("bts", TrianglesBottomBannerPattern())
	registerBannerPattern("tts", TrianglesTopBannerPattern())
	registerBannerPattern("glb", GlobeBannerPattern())
	registerBannerPattern("pig", PiglinBannerPattern())
}

// registerBannerPattern registers a banner pattern with the ID passed.
func registerBannerPattern(id string, pattern BannerPatternType) {
	bannerPatternsMap[id] = pattern
	bannerPatternIDs[pattern] = id
}

// bannerPatternByID returns a banner pattern by the ID it was registered with.
func bannerPatternByID(id string) BannerPatternType {
	b, ok := bannerPatternsMap[id]
	if !ok {
		panic("should never happen")
	}
	return b
}

// bannerPatternID returns the ID a banner pattern was registered with.
func bannerPatternID(pattern BannerPatternType) string {
	id, ok := bannerPatternIDs[pattern]
	if !ok {
		panic("should never happen")
	}
	return id
}
