package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

type (
	// Infested is a block that hides a silverfish. It looks identical to its non-infested counterpart, but is broken
	// twice as fast.
	Infested struct {
		solid
		bassDrum

		// Type is the type of infested block.
		Type InfestedType
	}

	// InfestedType represents a type of infested block.
	InfestedType struct {
		infested
	}

	infested uint8
)

// StoneInfested is the stone variant of infested blocks.
func StoneInfested() InfestedType {
	return InfestedType{0}
}

// CobblestoneInfested is the cobblestone variant of infested blocks.
func CobblestoneInfested() InfestedType {
	return InfestedType{1}
}

// StoneBricksInfested is the stone bricks variant of infested blocks.
func StoneBricksInfested() InfestedType {
	return InfestedType{2}
}

// MossyStoneBricksInfested is the mossy stone bricks variant of infested blocks.
func MossyStoneBricksInfested() InfestedType {
	return InfestedType{3}
}

// CrackedStoneBricksInfested is the cracked stone bricks variant of infested blocks.
func CrackedStoneBricksInfested() InfestedType {
	return InfestedType{4}
}

// ChiseledStoneBricksInfested is the chiseled stone bricks variant of infested blocks.
func ChiseledStoneBricksInfested() InfestedType {
	return InfestedType{5}
}

// DeepslateInfested is the deepslate variant of infested blocks.
func DeepslateInfested() InfestedType {
	return InfestedType{6}
}

// BreakInfo ...
func (i Infested) BreakInfo() BreakInfo {
	// Infested blocks have half the hardness of their normal counterparts. In vanilla, they drop nothing
	// and spawn a silverfish, but since mobs aren't implemented yet, we just drop nothing.
	return newBreakInfo(0.75, alwaysHarvestable, nothingEffective, nil).withBlastResistance(0.75)
}

// EncodeItem ...
func (i Infested) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_" + i.Type.String(), 0
}

// EncodeBlock ...
func (i Infested) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_" + i.Type.String(), nil
}

// Normal returns the normal block variant of the infested type.
func (i infested) Normal() world.Block {
	switch i {
	case 0:
		return Stone{}
	case 1:
		return Cobblestone{}
	case 2:
		return StoneBricks{Type: NormalStoneBricks()}
	case 3:
		return StoneBricks{Type: MossyStoneBricks()}
	case 4:
		return StoneBricks{Type: CrackedStoneBricks()}
	case 5:
		return StoneBricks{Type: ChiseledStoneBricks()}
	case 6:
		return Deepslate{Type: NormalDeepslate()}
	}
	panic("unknown infested type")
}

// Uint8 returns the infested type as a uint8.
func (i infested) Uint8() uint8 {
	return uint8(i)
}

// Name ...
func (i infested) Name() string {
	switch i {
	case 0:
		return "Stone"
	case 1:
		return "Cobblestone"
	case 2:
		return "Stone Bricks"
	case 3:
		return "Mossy Stone Bricks"
	case 4:
		return "Cracked Stone Bricks"
	case 5:
		return "Chiseled Stone Bricks"
	case 6:
		return "Deepslate"
	}
	panic("unknown infested type")
}

// String ...
func (i infested) String() string {
	switch i {
	case 0:
		return "stone"
	case 1:
		return "cobblestone"
	case 2:
		return "stone_bricks"
	case 3:
		return "mossy_stone_bricks"
	case 4:
		return "cracked_stone_bricks"
	case 5:
		return "chiseled_stone_bricks"
	case 6:
		return "deepslate"
	}
	panic("unknown infested type")
}

// InfestedTypes ...
func InfestedTypes() []InfestedType {
	return []InfestedType{
		StoneInfested(),
		CobblestoneInfested(),
		StoneBricksInfested(),
		MossyStoneBricksInfested(),
		CrackedStoneBricksInfested(),
		ChiseledStoneBricksInfested(),
		DeepslateInfested(),
	}
}

// allInfested ...
func allInfested() (s []world.Block) {
	for _, t := range InfestedTypes() {
		s = append(s, Infested{Type: t})
	}
	return
}
