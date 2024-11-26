package block

// DeepslateType represents a type of deepslate.
type DeepslateType struct {
	deepslate
}

type deepslate uint8

// NormalDeepslate is the normal variant of deepslate.
func NormalDeepslate() DeepslateType {
	return DeepslateType{0}
}

// CobbledDeepslate is the cobbled variant of deepslate.
func CobbledDeepslate() DeepslateType {
	return DeepslateType{1}
}

// PolishedDeepslate is the polished variant of deepslate.
func PolishedDeepslate() DeepslateType {
	return DeepslateType{2}
}

// ChiseledDeepslate is the chiseled variant of deepslate.
func ChiseledDeepslate() DeepslateType {
	return DeepslateType{3}
}

// Uint8 returns the deepslate type as a uint8.
func (s deepslate) Uint8() uint8 {
	return uint8(s)
}

// Name ...
func (s deepslate) Name() string {
	switch s {
	case 0:
		return "Deepslate"
	case 1:
		return "Cobbled Deepslate"
	case 2:
		return "Polished Deepslate"
	case 3:
		return "Chiseled Deepslate"
	}
	panic("unknown deepslate type")
}

// String ...
func (s deepslate) String() string {
	switch s {
	case 0:
		return "deepslate"
	case 1:
		return "cobbled_deepslate"
	case 2:
		return "polished_deepslate"
	case 3:
		return "chiseled_deepslate"
	}
	panic("unknown deepslate type")
}

// DeepslateTypes ...
func DeepslateTypes() []DeepslateType {
	return []DeepslateType{NormalDeepslate(), CobbledDeepslate(), PolishedDeepslate(), ChiseledDeepslate()}
}
