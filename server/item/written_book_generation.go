package item

// WrittenBookGeneration represents a WrittenBook generation.
type WrittenBookGeneration struct {
	generation
}

type generation uint8

// OriginalGeneration is the original WrittenBook.
func OriginalGeneration() WrittenBookGeneration {
	return WrittenBookGeneration{0}
}

// CopyGeneration is a copy of the original WrittenBook.
func CopyGeneration() WrittenBookGeneration {
	return WrittenBookGeneration{1}
}

// CopyOfCopyGeneration is a copy of a copy of the original WrittenBook.
func CopyOfCopyGeneration() WrittenBookGeneration {
	return WrittenBookGeneration{2}
}

// Uint8 returns the generation as a uint8.
func (g generation) Uint8() uint8 {
	return uint8(g)
}

// String ...
func (g generation) String() string {
	switch g {
	case 0:
		return "original"
	case 1:
		return "copy of original"
	case 2:
		return "copy of copy"
	}
	panic("unknown written book generation")
}
