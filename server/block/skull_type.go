package block

import (
	"fmt"
)

// SkullType represents a mob variant of a skull.
type SkullType struct {
	skull
}

// SkeletonSkull returns the skull variant for skeletons.
func SkeletonSkull() SkullType {
	return SkullType{0}
}

// WitherSkeletonSkull returns the skull variant for wither skeletons.
func WitherSkeletonSkull() SkullType {
	return SkullType{1}
}

// ZombieHead returns the skull variant for zombies.
func ZombieHead() SkullType {
	return SkullType{2}
}

// PlayerHead returns the skull variant for players.
func PlayerHead() SkullType {
	return SkullType{3}
}

// CreeperHead returns the skull variant for creepers.
func CreeperHead() SkullType {
	return SkullType{4}
}

// DragonHead returns the skull variant for ender dragons.
func DragonHead() SkullType {
	return SkullType{5}
}

// SkullTypes returns all variants of skulls.
func SkullTypes() []SkullType {
	return []SkullType{SkeletonSkull(), WitherSkeletonSkull(), ZombieHead(), PlayerHead(), CreeperHead(), DragonHead()}
}

type skull uint8

// Uint8 ...
func (s skull) Uint8() uint8 {
	return uint8(s)
}

// Name ...
func (s skull) Name() string {
	switch s {
	case 0:
		return "Skeleton Skull"
	case 1:
		return "Wither Skeleton Skull"
	case 2:
		return "Zombie Head"
	case 3:
		return "Player Head"
	case 4:
		return "Creeper Head"
	case 5:
		return "Dragon Head"
	}
	panic("unknown skull type")
}

// FromString ...
func (s skull) FromString(str string) (interface{}, error) {
	switch str {
	case "skeleton":
		return SkeletonSkull(), nil
	case "wither_skeleton":
		return WitherSkeletonSkull(), nil
	case "zombie":
		return ZombieHead(), nil
	case "player":
		return PlayerHead(), nil
	case "creeper":
		return CreeperHead(), nil
	case "dragon":
		return DragonHead(), nil
	}
	return nil, fmt.Errorf("unexpected skull type '%v', expecting one of 'skeleton', 'wither_skeleton', 'zombie', 'creeper', or 'dragon'", str)
}

// String ...
func (s skull) String() string {
	switch s {
	case 0:
		return "skeleton"
	case 1:
		return "wither_skeleton"
	case 2:
		return "zombie"
	case 3:
		return "player"
	case 4:
		return "creeper"
	case 5:
		return "dragon"
	}
	panic("unknown skull type")
}
