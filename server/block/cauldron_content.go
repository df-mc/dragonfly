package block

// CauldronContent is the material stored in a cauldron. Water may additionally
// contain a dye or a potion, which are stored in the cauldron's block entity.
type CauldronContent struct {
	cauldronContent
}

type cauldronContent uint8

// WaterCauldronContent is the default content of a cauldron.
func WaterCauldronContent() CauldronContent { return CauldronContent{0} }

// LavaCauldronContent returns lava cauldron content.
func LavaCauldronContent() CauldronContent { return CauldronContent{1} }

// PowderSnowCauldronContent returns powder snow cauldron content.
func PowderSnowCauldronContent() CauldronContent { return CauldronContent{2} }

// Uint8 returns the content as a uint8.
func (c cauldronContent) Uint8() uint8 { return uint8(c) }

// String returns the block state representation of the content.
func (c cauldronContent) String() string {
	switch c {
	case 0:
		return "water"
	case 1:
		return "lava"
	case 2:
		return "powder_snow"
	}
	panic("unknown cauldron content")
}
