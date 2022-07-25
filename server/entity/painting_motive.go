package entity

type PaintingMotive struct {
	paintingMotive
}

// AlbanPainting is the alban motive of a painting.
func AlbanPainting() PaintingMotive {
	return PaintingMotive{0}
}

// AztecPainting is the aztec motive of a painting.
func AztecPainting() PaintingMotive {
	return PaintingMotive{1}
}

// Aztec2Painting is the aztec2 motive of a painting.
func Aztec2Painting() PaintingMotive {
	return PaintingMotive{2}
}

// BombPainting is the bomb motive of a painting.
func BombPainting() PaintingMotive {
	return PaintingMotive{3}
}

// KebabPainting is the kebab motive of a painting.
func KebabPainting() PaintingMotive {
	return PaintingMotive{4}
}

// PlantPainting is the plant motive of a painting.
func PlantPainting() PaintingMotive {
	return PaintingMotive{5}
}

// WastelandPainting is the wasteland motive of a painting.
func WastelandPainting() PaintingMotive {
	return PaintingMotive{6}
}

// CourbetPainting is the courbet motive of a painting.
func CourbetPainting() PaintingMotive {
	return PaintingMotive{7}
}

// PoolPainting is the pool motive of a painting.
func PoolPainting() PaintingMotive {
	return PaintingMotive{8}
}

// SeaPainting is the sea motive of a painting.
func SeaPainting() PaintingMotive {
	return PaintingMotive{9}
}

// CreebetPainting is the creebet motive of a painting.
func CreebetPainting() PaintingMotive {
	return PaintingMotive{10}
}

// SunsetPainting is the sunset motive of a painting.
func SunsetPainting() PaintingMotive {
	return PaintingMotive{11}
}

// GrahamPainting is the graham motive of a painting.
func GrahamPainting() PaintingMotive {
	return PaintingMotive{12}
}

// WandererPainting is the wanderer motive of a painting.
func WandererPainting() PaintingMotive {
	return PaintingMotive{13}
}

// BustPainting is the bust motive of a painting.
func BustPainting() PaintingMotive {
	return PaintingMotive{14}
}

// MatchPainting is the match motive of a painting.
func MatchPainting() PaintingMotive {
	return PaintingMotive{15}
}

// SkullAndRosesPainting is the skull and roses motive of a painting.
func SkullAndRosesPainting() PaintingMotive {
	return PaintingMotive{16}
}

// StagePainting is the stage motive of a painting.
func StagePainting() PaintingMotive {
	return PaintingMotive{17}
}

// VoidPainting is the void motive of a painting.
func VoidPainting() PaintingMotive {
	return PaintingMotive{18}
}

// WitherPainting is the wither motive of a painting.
func WitherPainting() PaintingMotive {
	return PaintingMotive{19}
}

// EarthPainting is the earth motive of a painting.
func EarthPainting() PaintingMotive {
	return PaintingMotive{20}
}

// FirePainting is the fire motive of a painting.
func FirePainting() PaintingMotive {
	return PaintingMotive{21}
}

// WaterPainting is the water motive of a painting.
func WaterPainting() PaintingMotive {
	return PaintingMotive{22}
}

// WindPainting is the wind motive of a painting.
func WindPainting() PaintingMotive {
	return PaintingMotive{23}
}

// FightersPainting is the fighters motive of a painting.
func FightersPainting() PaintingMotive {
	return PaintingMotive{24}
}

// DonkeyKongPainting is the donkey kong motive of a painting.
func DonkeyKongPainting() PaintingMotive {
	return PaintingMotive{25}
}

// SkeletonPainting is the skeleton motive of a painting.
func SkeletonPainting() PaintingMotive {
	return PaintingMotive{26}
}

// BurningSkullPainting is the burning skull motive of a painting.
func BurningSkullPainting() PaintingMotive {
	return PaintingMotive{27}
}

// PigScenePainting is the pig scene motive of a painting.
func PigScenePainting() PaintingMotive {
	return PaintingMotive{28}
}

// PointerPainting is the pointer motive of a painting.
func PointerPainting() PaintingMotive {
	return PaintingMotive{29}
}

// PaintingMotives returns all the possible motives for a painting.
func PaintingMotives() []PaintingMotive {
	return []PaintingMotive{AlbanPainting(), AztecPainting(), Aztec2Painting(), BombPainting(), KebabPainting(),
		PlantPainting(), WastelandPainting(), CourbetPainting(), PoolPainting(), SeaPainting(), CreebetPainting(),
		SunsetPainting(), GrahamPainting(), WandererPainting(), BustPainting(), MatchPainting(), SkullAndRosesPainting(),
		StagePainting(), VoidPainting(), WitherPainting(), FightersPainting(), DonkeyKongPainting(), SkeletonPainting(),
		BurningSkullPainting(), PigScenePainting(), PointerPainting(),
	}
}

type paintingMotive uint8

// Size returns the size of the motive in the 2D axis.
func (p paintingMotive) Size() (int, int) {
	if p.Uint8() < 7 {
		return 1, 1
	} else if p.Uint8() < 12 {
		return 2, 1
	} else if p.Uint8() < 14 {
		return 1, 2
	} else if p.Uint8() < 24 {
		return 2, 2
	} else if p.Uint8() < 25 {
		return 4, 2
	} else if p.Uint8() < 27 {
		return 4, 3
	} else if p.Uint8() < 30 {
		return 4, 4
	}
	panic("unknown painting type")
}

// Uint8 ...
func (p paintingMotive) Uint8() uint8 {
	return uint8(p)
}

// String ...
func (p paintingMotive) String() string {
	switch p.Uint8() {
	case 0:
		return "Alban"
	case 1:
		return "Aztec"
	case 2:
		return "Aztec2"
	case 3:
		return "Bomb"
	case 4:
		return "Kebab"
	case 5:
		return "Plant"
	case 6:
		return "Wasteland"
	case 7:
		return "Courbet"
	case 8:
		return "Pool"
	case 9:
		return "Sea"
	case 10:
		return "Creebet"
	case 11:
		return "Sunset"
	case 12:
		return "Graham"
	case 13:
		return "Wanderer"
	case 14:
		return "Bust"
	case 15:
		return "Match"
	case 16:
		return "SkullAndRoses"
	case 17:
		return "Stage"
	case 18:
		return "Void"
	case 19:
		return "Wither"
	case 20:
		return "Earth"
	case 21:
		return "Fire"
	case 22:
		return "Water"
	case 23:
		return "Wind"
	case 24:
		return "Fighters"
	case 25:
		return "DonkeyKong"
	case 26:
		return "Skeleton"
	case 27:
		return "BurningSkull"
	case 28:
		return "Pigscene"
	case 29:
		return "Pointer"
	}
	panic("unknown painting type")
}

func PaintingMotiveFromString(name string) PaintingMotive {
	switch name {
	case "Alban":
		return AlbanPainting()
	case "Aztec":
		return AztecPainting()
	case "Aztec2":
		return Aztec2Painting()
	case "Bomb":
		return BombPainting()
	case "Kebab":
		return KebabPainting()
	case "Plant":
		return PlantPainting()
	case "Wasteland":
		return WastelandPainting()
	case "Courbet":
		return CourbetPainting()
	case "Pool":
		return PoolPainting()
	case "Sea":
		return SeaPainting()
	case "Creebet":
		return CreebetPainting()
	case "Sunset":
		return SunsetPainting()
	case "Graham":
		return GrahamPainting()
	case "Wanderer":
		return WandererPainting()
	case "Bust":
		return BustPainting()
	case "Match":
		return MatchPainting()
	case "SkullAndRoses":
		return SkullAndRosesPainting()
	case "Stage":
		return StagePainting()
	case "Void":
		return VoidPainting()
	case "Wither":
		return WitherPainting()
	case "Earth":
		return EarthPainting()
	case "Fire":
		return FirePainting()
	case "Water":
		return WaterPainting()
	case "Wind":
		return WindPainting()
	case "Fighters":
		return FightersPainting()
	case "DonkeyKong":
		return DonkeyKongPainting()
	case "Skeleton":
		return SkeletonPainting()
	case "BurningSkull":
		return BurningSkullPainting()
	case "Pigscene":
		return PigScenePainting()
	case "Pointer":
		return PointerPainting()
	}
	panic("unknown painting type")
}
