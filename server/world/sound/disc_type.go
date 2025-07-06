package sound

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// DiscType represents the type of music disc. Typically, Minecraft has a total of 15 music discs.
type DiscType struct {
	disc
}

// Disc13 returns the music disc "13".
func Disc13() DiscType {
	return DiscType{0}
}

// DiscCat returns the music disc "cat".
func DiscCat() DiscType {
	return DiscType{1}
}

// DiscBlocks returns the music disc "blocks".
func DiscBlocks() DiscType {
	return DiscType{2}
}

// DiscChirp returns the music disc "chirp".
func DiscChirp() DiscType {
	return DiscType{3}
}

// DiscFar returns the music disc "far".
func DiscFar() DiscType {
	return DiscType{4}
}

// DiscMall returns the music disc "mall".
func DiscMall() DiscType {
	return DiscType{5}
}

// DiscMellohi returns the music disc "mellohi".
func DiscMellohi() DiscType {
	return DiscType{6}
}

// DiscStal returns the music disc "stal".
func DiscStal() DiscType {
	return DiscType{7}
}

// DiscStrad returns the music disc "strad".
func DiscStrad() DiscType {
	return DiscType{8}
}

// DiscWard returns the music disc "ward".
func DiscWard() DiscType {
	return DiscType{9}
}

// Disc11 returns the music disc "11".
func Disc11() DiscType {
	return DiscType{10}
}

// DiscWait returns the music disc "wait".
func DiscWait() DiscType {
	return DiscType{11}
}

// DiscOtherside returns the music disc "otherside".
func DiscOtherside() DiscType {
	return DiscType{12}
}

// DiscPigstep returns the music disc "Pigstep".
func DiscPigstep() DiscType {
	return DiscType{13}
}

// Disc5 returns the music disc "5".
func Disc5() DiscType {
	return DiscType{14}
}

// DiscRelic returns the music disc "Relic".
func DiscRelic() DiscType {
	return DiscType{15}
}

// DiscCreator returns the music disc "Creator".
func DiscCreator() DiscType {
	return DiscType{16}
}

// DiscCreatorMusicBox returns the music disc "Creator (Music Box)".
func DiscCreatorMusicBox() DiscType {
	return DiscType{17}
}

// DiscPrecipice returns the music disc "Precipice".
func DiscPrecipice() DiscType {
	return DiscType{18}
}

// DiscTears returns the music disc "Tears".
func DiscTears() DiscType {
	return DiscType{19}
}

// DiscLavaChicken returns the music disc "Lava Chicken".
func DiscLavaChicken() DiscType {
	return DiscType{20}
}

// MusicDiscs returns a list of all existing music discs.
func MusicDiscs() []DiscType {
	return []DiscType{
		Disc13(), DiscCat(), DiscBlocks(), DiscChirp(), DiscFar(), DiscMall(), DiscMellohi(), DiscStal(),
		DiscStrad(), DiscWard(), Disc11(), DiscWait(), DiscOtherside(), DiscPigstep(), Disc5(), DiscRelic(),
		DiscCreator(), DiscCreatorMusicBox(), DiscPrecipice(), DiscTears(), DiscLavaChicken(),
	}
}

// disc is the underlying value of a DiscType struct.
type disc uint8

// Uint8 converts the disc to an integer that uniquely identifies its type.
func (d disc) Uint8() uint8 {
	return uint8(d)
}

// String ...
func (d disc) String() string {
	switch d {
	case 0:
		return "13"
	case 1:
		return "cat"
	case 2:
		return "blocks"
	case 3:
		return "chirp"
	case 4:
		return "far"
	case 5:
		return "mall"
	case 6:
		return "mellohi"
	case 7:
		return "stal"
	case 8:
		return "strad"
	case 9:
		return "ward"
	case 10:
		return "11"
	case 11:
		return "wait"
	case 12:
		return "otherside"
	case 13:
		return "pigstep"
	case 14:
		return "5"
	case 15:
		return "relic"
	case 16:
		return "creator"
	case 17:
		return "creator_music_box"
	case 18:
		return "precipice"
	case 19:
		return "tears"
	case 20:
		return "lava_chicken"
	}
	panic("unknown record type")
}

// DisplayName ...
func (d disc) DisplayName() string {
	switch d {
	case 17:
		return "Creator (Music Box)"
	}
	if d > 12 {
		return cases.Title(language.English, cases.Compact).String(d.String())
	}
	return d.String()
}

// Author ...
func (d disc) Author() string {
	if d <= 11 {
		return "C418"
	}
	switch d {
	case 12, 13, 16, 17:
		return "Lena Raine"
	case 14:
		return "Samuel Åberg"
	case 15, 18:
		return "Aaron Cherof"
	case 19:
		return "Amos Roddy"
	case 20:
		return "Hyper Potions"
	}
	panic("unknown record type")
}
